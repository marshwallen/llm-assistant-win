// Get-WinEvent文档：
// https://learn.microsoft.com/zh-cn/powershell/module/microsoft.powershell.diagnostics/get-winevent?view=powershell-7.5

package tools

import (
	"fmt"
	"log"
	"winds-assistant/utils"
)

// QueryEvents 根据给定的事件查询条件 q，查询指定日志中的事件。
func QueryEvents(q map[string]interface{}) (string, error) {
    logName, _ := q["logName"].(string)
    _startTime, _ := q["startTime"].(float64)
    _maxEvents, _ := q["maxEvents"].(float64)

    startTime := int(_startTime)
    maxEvents := int(_maxEvents)

    print(startTime, maxEvents)
    startTimeFormat := fmt.Sprintf("(Get-Date).AddDays(%v)", -startTime)
    // 安全创建命令对象（拆分命令和参数）
    out, _, err := utils.RunCommand("PowerShell", "-Command", 
        "Get-WinEvent", 
        "-FilterHashtable",
        fmt.Sprintf("@{ LogName='%s'; StartTime=%s }", logName, startTimeFormat,),
        "-MaxEvents", fmt.Sprint(maxEvents),
        "| Out-String -Width 4096",
    )
    
    // 执行并检查错误
    if err != nil {
        log.Printf("[%s] Query Error: %v\n",
        logName,
            err,
        );return "", err
    }

    return out, nil
}
