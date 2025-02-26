// Get-WinEvent文档：
// https://learn.microsoft.com/zh-cn/powershell/module/microsoft.powershell.diagnostics/get-winevent?view=powershell-7.5

package tools

import (
    "winds-assistant/utils"
    "fmt"
    "log"
)

// 构造 Get-WinEvent 查询体
type EventQuery struct {
    LogName   string        // 日志类型 (Application, Security, System, etc.)
    StartTime  int          // 起始时间 (负数，代表往前推多少天)
    MaxEvents  int          // 最大事件数
}

// QueryEvents 根据给定的事件查询条件 q，查询指定日志中的事件。
func QueryEvents(q EventQuery) (map[string]string, error) {
    res := make(map[string]string)
    
    startTime := fmt.Sprintf("(Get-Date).AddDays(%v)", q.StartTime)
    // 安全创建命令对象（拆分命令和参数）
    out, _, err := utils.RunCommand("PowerShell", "-Command", 
        "Get-WinEvent", 
        "-FilterHashtable",
        fmt.Sprintf("@{ LogName='%s'; StartTime=%s }", q.LogName, startTime,),
        "-MaxEvents", fmt.Sprint(q.MaxEvents),
        "| Out-String -Width 4096",
    )
    
    // 执行并检查错误
    if err != nil {
        log.Printf("[%s] Query Error: %v\n",
        q.LogName,
            err,
        );return nil, err
    }
    res[q.LogName] = out
    
    return res, nil
}
