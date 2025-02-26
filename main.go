package main

import (
    "winds-assistant/ui"
    "winds-assistant/workers"
    "context"
)

func main() {
    ctx := context.Background()
	stop := workers.MonitorSys(ctx)
	ui.StartAPP()
	stop()
}