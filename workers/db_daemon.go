package workers

import (
	"context"
	"log"
	"time"
	"sync"
	"winds-assistant/utils"
	"strconv"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"winds-assistant/tools"
	"fmt"
	"os"
	"strings"
)

type MetricData struct {
	CPU  cpu.InfoStat
	Mem  *mem.VirtualMemoryStat
	Disk *disk.UsageStat
	GPU  []tools.GPUStats
	Time int64
}

type DailyWriter struct {
    writers    map[string]*utils.CSVWriter
    currentDay string
    mu         sync.Mutex
}

// NewDailyWriter 创建一个新的 DailyWriter 实例。(单例模式)
func NewDailyWriter() *DailyWriter {
    return &DailyWriter{
        writers:    make(map[string]*utils.CSVWriter),
        currentDay: time.Now().Local().Format(dateFormat),
    }
}

const (
	collectInterval = 5 * time.Second
	storeInterval   = 5 * time.Second
	bufferSize      = 1024
	dateFormat = "20060102"
)

var dailyWriter = NewDailyWriter()

// 启动系统监控并返回停止函数
func MonitorSys(ctx context.Context) (stopFunc func()) {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	// 初始化数据存放路径
	if err := utils.EnsureDir("data/"); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	dataChan := make(chan MetricData, bufferSize)

	// 启动采集协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		collectMetrics(ctx, dataChan)
	}()

	// 启动存储协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		storeMetrics(ctx, dailyWriter.writers, dataChan)
	}()

	return func() {
		cancel()  // 发送停止信号
		wg.Wait() // 等待所有协程退出
	}
}

// 定期收集系统指标并将其发送到指定的通道。
func collectMetrics(ctx context.Context, dataChan chan<- MetricData) {
	ticker := time.NewTicker(collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			close(dataChan) // 关闭通道通知存储协程
			return
		case <-ticker.C:
			data, err := gatherSystemMetrics()
			if err != nil {
				log.Printf("Metrics collection failed: %v", err)
				continue
			}

			select {
			case dataChan <- data:
			default:
				log.Println("Metrics buffer full, discarding data")
			}
		}
	}
}

// 收集系统的各项指标数据，包括 CPU、内存、磁盘和 GPU 信息。
func gatherSystemMetrics() (MetricData, error) {
	cpuInfo, err := tools.GetCPUInfo()
	if err != nil {
		return MetricData{}, err
	}

	memInfo, err := tools.GetMemInfo()
	if err != nil {
		return MetricData{}, err
	}

	diskInfo, err := tools.GetDiskInfo()
	if err != nil {
		return MetricData{}, err
	}

	gpuInfo, err := tools.GetNVGPUInfo()
	if err != nil {
		return MetricData{}, err
	}

	return MetricData{
		CPU:  cpuInfo,
		Mem:  memInfo,
		Disk: diskInfo,
		GPU:  gpuInfo,
		Time: time.Now().Local().Unix(),
	}, nil
}

// 持续接收 MetricData 并定期将其存储到 CSV 文件中。
func storeMetrics(ctx context.Context, writers map[string]*utils.CSVWriter, dataChan <-chan MetricData) {
	ticker := time.NewTicker(storeInterval)
	defer ticker.Stop()

	var batch []MetricData
	for {
		select {
		case <-ctx.Done():
			flushRemainingData(batch)
			// 记得关闭所有 writer
			for _, writer := range writers {
				writer.Close()
			}
			return
		case data, ok := <-dataChan:
			if !ok {
				flushRemainingData(batch)
				return
			}
			batch = append(batch, data)
		case <-ticker.C:
			if len(batch) > 0 {
				writeBatchToCSV(batch)
				batch = nil
			}
		}
	}
}

// 将一批 MetricData 写入 CSV 文件。
func writeBatchToCSV(batch []MetricData) {
    for _, data := range batch {
        dailyWriter.WriteMetric(data.Time, "cpu_clock", data.CPU.Mhz, "MHz")
        dailyWriter.WriteMetric(data.Time, "mem_used", float64(data.Mem.Used), "bytes")
        dailyWriter.WriteMetric(data.Time, "disk_used", float64(data.Disk.Used), "bytes")
        
        if len(data.GPU) > 0 {
            g := data.GPU[0]
            dailyWriter.WriteMetric(data.Time, "gpu_util", float64(g.Utilization), "%")
            dailyWriter.WriteMetric(data.Time, "gpu_mem_used", float64(g.MemUsed), "MB")
            dailyWriter.WriteMetric(data.Time, "gpu_mem_clock", float64(g.MemClock), "MHz")
            dailyWriter.WriteMetric(data.Time, "gpu_core_clock", float64(g.CoreClock), "MHz")
            dailyWriter.WriteMetric(data.Time, "gpu_temp", float64(g.Temperature), "°C")
            dailyWriter.WriteMetric(data.Time, "gpu_power", float64(g.PowerDraw), "W")
        }
    }
    
    // 批量刷新
    dailyWriter.mu.Lock()
    defer dailyWriter.mu.Unlock()
    for _, w := range dailyWriter.writers {
        w.Flush()
    }
	// 删除旧文件（三天以前）
    dailyWriter.rotateFiles()
}

// 根据新的日期旋转文件，清理旧文件并重置写入器状态。
// 该方法会删除三天前的旧文件，关闭当前的写入器，并初始化新的写入器状态。
func (dw *DailyWriter) rotateFiles() {
    // 清理旧文件
    oldDay := time.Now().AddDate(0, 0, -3).Local().Format(dateFormat)

	files, err := os.ReadDir("data/")
    if err != nil {
        log.Fatalf("%v", err)
    }

	for _, file := range files {
        if file.IsDir() {
            continue // 跳过子目录
        }
        
        // 只处理 CSV 文件
        if !strings.HasSuffix(file.Name(), ".csv") {
            continue
        }
        
		_f := strings.Split(file.Name(), "_")
		_c := strings.Compare(strings.Split(_f[len(_f)-1], ".")[0], oldDay)
		if _c<=0 {
			oldFile := fmt.Sprintf("data/%s", file.Name())
			if _, err := os.Stat(oldFile); err == nil {
				os.Remove(oldFile)
			}
		}
    }
}

// 将指定时间戳、名称、值和单位写入 CSV 文件
func (dw *DailyWriter) WriteMetric(timestamp int64, name string, value float64, unit string) {
    dw.mu.Lock()
    defer dw.mu.Unlock()

	// 更新 dailyWriter 的当前日期
	currentDay := time.Unix(timestamp, 0).Format(dateFormat)
	dw.currentDay = currentDay

    // 获取或创建写入器
    writer, exists := dw.writers[name]
    if !exists {
        fileName := fmt.Sprintf("data/%s_%s.csv", name, dw.currentDay)
        var err error
        if writer, err = utils.NewCSVWriter(fileName); err != nil {
            log.Printf("创建写入器失败: %v", err)
            return
        }
        dw.writers[name] = writer
        writer.Write([]string{"timestamp", "time", "name", "value", "unit", "source"}) // 写入表头
    }

    // 构造记录
    record := []string{
		fmt.Sprint(timestamp),
        time.Unix(timestamp, 0).Format(time.RFC3339Nano),
        name,
        strconv.FormatFloat(value, 'f', 2, 64),
        unit,
        "system",
    }

    if err := writer.Write(record); err != nil {
        log.Printf("写入失败: %v", err)
    }
}

// 处理剩余数据
func flushRemainingData(batch []MetricData) {
    if len(batch) > 0 {
        log.Printf("Flushing %d pending metrics", len(batch))
        writeBatchToCSV(batch)
    }
}