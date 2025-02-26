package tools

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/disk"
	"winds-assistant/utils"
	"fmt"
	"strings"
)

type GPUStats struct {
	Index		   uint64  `json:"index"`		 // GPU序号
	Name		   string  `json:"name"`         // GPU名称
    Utilization    float64 `json:"gpu_util"`     // GPU利用率(%) 
    MemUsed        uint64  `json:"mem_used"`     // 显存使用量(MB)
    MemTotal       uint64  `json:"mem_total"`    // 显存总量(MB)
    CoreClock      uint64  `json:"core_clock"`   // 核心频率(MHz)
    MemClock       uint64  `json:"mem_clock"`    // 显存频率(MHz)
    Temperature    uint64  `json:"temp"`         // 温度(℃)
    PowerDraw      float64 `json:"power"`        // 实时功耗(W)
    Vendor         string  `json:"vendor"`       // 厂商(NVIDIA)
}

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
func GetNVGPUInfo() ([]GPUStats, error){
	out, _, err := utils.RunCommand("nvidia-smi", 
        "--query-gpu=index,name,utilization.gpu,memory.used,memory.total,clocks.current.graphics,clocks.current.memory,temperature.gpu,power.draw",
        "--format=csv,noheader,nounits")
	if err != nil {
		return nil, fmt.Errorf("执行nvidia-smi失败: %v\n输出: %s", err, out)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
    var stats []GPUStats

    for _, line := range lines {
        fields := strings.Split(line, ", ")
        if len(fields) < 8 { // 包含index后字段数应为8
            continue // 跳过格式异常行
        }

        stat := GPUStats{
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