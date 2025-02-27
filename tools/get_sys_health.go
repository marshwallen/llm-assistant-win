package tools

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/disk"
	"winds-assistant/utils"
	"fmt"
	"strings"
	"winds-assistant/common"
	"time"
	"os"
	"encoding/csv"
)

// 获取CPU信息
func GetCPUInfo() (cpu.InfoStat, error) {
	var ret cpu.InfoStat
	percent, err := cpu.Info()
	if err != nil {
		return ret, err
	}
	return percent[0], nil
}

// 获取NVIDIA GPU信息
func GetNVGPUInfo() ([]common.GPUStats, error){
	out, _, err := utils.RunCommand("nvidia-smi", 
        "--query-gpu=index,name,utilization.gpu,memory.used,memory.total,clocks.current.graphics,clocks.current.memory,temperature.gpu,power.draw",
        "--format=csv,noheader,nounits")
	if err != nil {
		return nil, fmt.Errorf("执行nvidia-smi失败: %v\n输出: %s", err, out)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
    var stats []common.GPUStats

    for _, line := range lines {
        fields := strings.Split(line, ", ")
        if len(fields) < 8 { // 包含index后字段数应为8
            continue // 跳过格式异常行
        }

        stat := common.GPUStats{
			Index:			utils.ParseUint(fields[0]),    // 字段0: index
			Name:			fields[1],                     // 字段1: name
            Utilization:    utils.ParseFloat(fields[2]),   // 字段2: utilization.gpu
            MemUsed:        utils.ParseUint(fields[3]),    // 字段3: memory.used
            MemTotal:       utils.ParseUint(fields[4]),    // 字段4: memory.total
            CoreClock:      utils.ParseUint(fields[5]),    // 字段5: core_clock
            MemClock:       utils.ParseUint(fields[6]),    // 字段6: mem_clock
            Temperature:    utils.ParseUint(fields[7]),    // 字段7: temperature
            PowerDraw:      utils.ParseFloat(fields[8]),   // 字段8: power
            Vendor:         "NVIDIA",
        }
        stats = append(stats, stat)
    }

	return stats, nil
}

// 获取内存利用率
func GetMemInfo() (*mem.VirtualMemoryStat, error) {
	var ret *mem.VirtualMemoryStat
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return ret, err
	}
	return memInfo, nil
}

// 获取磁盘利用率
func GetDiskInfo() (*disk.UsageStat, error) {
	var ret *disk.UsageStat
	usage, err := disk.Usage("C://")
	if err != nil {
		return ret, err
	}
	return usage, nil
}

var metrics = map[string]string{
	"cpu_clock": "MHz", 
	"mem_used": "bytes", 
	"disk_used": "bytes",
	"gpu_util": "%", 
	"gpu_mem_used": "MB", 
	"gpu_mem_clock": "MHz",
	"gpu_core_clock": "MHz", 
	"gpu_temp": "°C",
	"gpu_power": "W",
}

// 读取 CSV 文件的后 N 行中的某一列的值
func readCSVLastNColumn(filename string, columnIndex int, n int) ([]string, error) {
	// 打开 CSV 文件
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// 创建 CSV 读取器
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv failed: %w", err)
	}

	// 检查列索引是否有效
	if columnIndex < 0 || columnIndex >= len(rows[0]) {
		return nil, fmt.Errorf("invalid column index: %d", columnIndex)
	}

	// 计算起始行
	start := len(rows) - n
	if start < 0 {
		start = 0
	}

	// 提取指定列的值
	var values []string
	for i := len(rows) - 1; i >= start; i-- {
		values = append(values, rows[i][columnIndex])
	}
	return values, nil
}

const dateFormat = "20060102"

func GetSysHealthData(minutes int) (out string, err error){
	cpuInfo , _ := GetCPUInfo()
	gpuInfo, _ := GetNVGPUInfo()
	memInfo, _ := GetMemInfo()
	diskInfo, _ := GetDiskInfo()

	out += "CPU 当前状态: " + fmt.Sprintf("%+v\n", cpuInfo)
	out += "GPU 当前状态: " + fmt.Sprintf("%+v\n", gpuInfo)
	out += "MEM 当前状态: " + fmt.Sprintf("%+v\n", memInfo)
	out += "C:/ 当前状态: " + fmt.Sprintf("%+v\n", diskInfo)

	currentDay := time.Now().Local().Format(dateFormat)
	for metric, unit := range metrics {
		values, err := readCSVLastNColumn(fmt.Sprintf("data/%s_%s.csv", metric, currentDay), 3, minutes*12)
		if err != nil {
			return "", err
		}
		out += fmt.Sprintf("%s 利用情况趋势如下 (每个指标间隔 5s, 单位为 %s): %s\n", metric, unit, fmt.Sprint(values))
	}
	return
}