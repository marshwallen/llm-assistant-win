package ui

import (
	"container/list"
	"context"
	"fmt"
	"winds-assistant/common"
	"winds-assistant/utils"
	"winds-assistant/workers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// 初始化并启动应用程序，设置用户界面和后端处理
func StartAPP() {
    // **UI Settings**
    wApp := app.New()
    window := wApp.NewWindow(common.WIDGET_APP_NAME)
    window.Resize(fyne.NewSize(1024, 768))

    cfg, _ := utils.LoadLLMCfg()
    fastCliboard, _ := utils.ReadTxtFile("config/fast_cliboard.txt")

    var modelList []string
    if cfg.Default == "ollama" {
        modelList = workers.GetModelList(cfg.Backend[cfg.Default].BaseURL, window)
    }else{
        modelList = []string{cfg.Backend[cfg.Default].Model}
    }
    
    if len(modelList) > 0 {
        backend := cfg.Backend[cfg.Default]
        backend.Model = modelList[0]
        cfg.Backend[cfg.Default] = backend
    }

    ctx, cancel := context.WithCancel(context.Background())
    settings := common.Settings{
        BackendName:    cfg.Default,
        BackendCfg:     cfg.Backend[cfg.Default],
        ModelList:      modelList,
        CancelFunc:     cancel,
        DialogID:       GenerateID(),
        EnableAgent:    false,
        SysPrompt:      workers.SYSTEM_PROMPT_DEFAULT,
        Running:        false,
        FastCliboard:   fastCliboard,
    }
    
    // **Backend Settings**
    history := list.New()
    
    widgets := MainWidgets(window, history, &settings, cfg)
    widgets.InputEntry.OnSubmitted = func(text string) {
        // 提交前先检查模型存不存在
        if settings.BackendCfg.Model == "" {
            common.ShowErrorDialog(window, fmt.Errorf("error: model not found"))
            return
        }

        if settings.Running {
            common.ShowErrorDialog(window, fmt.Errorf("info: assistant is running, terminate it first"))
            return
        }
        UpdateHistory(history, common.LLMMessage{Role: "user", Content: text})
        widgets.InputEntry.SetText("")

        widgets.ChatChunk.Process(fmt.Sprintf("%s%s\n", common.CHAT_USER_INFO, text))
        widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderNextText())

        ctx, cancel = context.WithCancel(context.Background())
        settings.CancelFunc = cancel

        widgets.ChatChunk.Process(common.CHAT_ASSISTANT_INFO)
        widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderFinalText())
        widgets.ChatScroll.ScrollToBottom()

        if settings.EnableAgent{
            go ProcessStreamWithTools(ctx, &settings, widgets, history)
        }else{
            go ProcessStream(ctx, &settings, widgets, history)
        }
    }

    // **APP Start**
    window.SetContent(widgets.MainSplit)
    window.ShowAndRun()
}



