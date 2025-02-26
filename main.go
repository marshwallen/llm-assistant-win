package main

import (
    "winds-assistant/ui"
)

func main() {
    // query := tools.EventQuery{
    //     LogName:  "Application",
    //     StartTime: -1,
    //     MaxEvents: 50,
    // }

    // res, err := tools.QueryEvents(query)
    // if err != nil {
    //     log.Fatal("查询失败:", err)
    // }

    // for logName, events := range res {
    //     fmt.Println(logName, events)
    // }

    // ctx := context.Background()
	// stop := workers.MonitorSys(ctx)

    // tools.SaveFileTree("E:/", 3, 10, false)

	ui.StartAPP()
	// **停止监控
	// stop()

    // // 调用示例
    // workers.StreamChat("ollama")

}