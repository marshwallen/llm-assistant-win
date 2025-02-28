package tools

import (
	"fmt"
    "github.com/shirou/gopsutil/v4/process"
)

// 获取系统中所有进程的关键信息
func GetSysProcess() (r []map[string]interface{}) {
	// 获取所有进程列表
    processes, _ := process.Processes()

	// 获取关于进程的关键信息，结合生产环境分析系统问题
    for _, p := range processes {
		pid := p.Pid						// 进程ID
		name, _ := p.Name()					// 进程名
		exePath, _ := p.Exe()         		// 执行路径
		// cmdline, _ := p.Cmdline()     		// 启动命令
		createTime, _ := p.CreateTime() 	// 创建时间戳（毫秒）
		status, _ := p.Status()       		// 运行状态（Running, Sleeping等）
		cpuPercent, _ := p.CPUPercent()		// CPU使用率
		memPercent, _ := p.MemoryPercent()	// 内存使用率
		memInfo, _ := p.MemoryInfo()  		// 内存详细信息（RSS/VMS）
		// username, _ := p.Username()   		// 运行用户（可能需要权限）
		numThreads, _ := p.NumThreads()		// 线程数
		ioCounters, _ := p.IOCounters()		// IO统计（读写次数、字节数）
		ppid, _ := p.Ppid()					// 父进程ID

		formatMessage := map[string]interface{}{
			"pid": 			pid,
			"name": 		name,
			"exePath": 		exePath,
			// "cmdline": 		cmdline,
			"createTime": 	createTime,
			"status": 		status,
			"cpuPercent": 	cpuPercent,
			"memPercent": 	memPercent,
			"memInfo": 		memInfo,
			// "username": 	username,
			"numThreads": 	numThreads,
			"ioCounters": 	ioCounters,
			"ppid": 		ppid,
		}
		r = append(r, formatMessage)
    }
	return
}

// 返回系统进程的字符串表示形式
func GetSysProcessStr() (r string) {
	raw := GetSysProcess()
	for _, v := range raw {
		r += fmt.Sprintf("%+v\n", v)
	}
	return
}