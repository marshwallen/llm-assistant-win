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

    cfg, _ := utils.LoadCfg()
    modelList := workers.GetModelList(cfg.Backend.OpenAI.URL, window)

    curModel := ""
    if len(modelList) > 0 {
        curModel = modelList[0]
    }

    ctx, cancel := context.WithCancel(context.Background())
    settings := common.Settings{
        URL:         cfg.Backend.OpenAI.URL,
        API_KEY:       cfg.Backend.OpenAI.API_KEY,
        Model:       curModel,
        ModelList:   modelList,
        CancelFunc:  cancel,
        DialogID:    GenerateID(),
        EnableAgent: false,
        SysPrompt:   workers.SYSTEM_PROMPT_DEFAULT,
        Running:     false,
    }
    
    // **Backend Settings**
    history := list.New()
    
    widgets := MainWidgets(window, &settings, history)
    widgets.InputEntry.OnSubmitted = func(text string) {
        // 提交前先检查模型存不存在
        if settings.Model == "" {
            common.ShowErrorDialog(window, fmt.Errorf("error: model not found"))
            return
        }

        if settings.Running {
            common.ShowErrorDialog(window, fmt.Errorf("info: assistant is running, terminate it first"))
            return
        }
        UpdateHistory(history, common.LLMMessage{Role: "User", Content: text})
        widgets.InputEntry.SetText("")

        widgets.ChatChunk.Process(fmt.Sprintf("%s%s\n", common.CHAT_USER_INFO, text))
        widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderNextText())

        ctx, cancel = context.WithCancel(context.Background())
        settings.CancelFunc = cancel

        widgets.ChatChunk.Process(common.CHAT_ASSISTANT_INFO)
        widgets.ChatDisplay.SetText(widgets.ChatChunk.RenderNextText())
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



