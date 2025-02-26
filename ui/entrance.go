package ui

import (
	"fmt"
	"winds-assistant/utils"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
    "winds-assistant/workers"
    "winds-assistant/common"
    "context"
)

// 初始化并启动应用程序，设置用户界面和后端处理
func StartAPP() {
    // **UI Settings**
    wApp := app.New()
    window := wApp.NewWindow(common.WIDGET_APP_NAME)
    window.Resize(fyne.NewSize(1024, 768))

    cfg, _ := utils.LoadCfg()
    modelList := workers.GetModelList(cfg)

    curModel := ""
    if len(modelList) > 0 {
        curModel = modelList[0]
    }

    ctx, cancel := context.WithCancel(context.Background())
    settings := common.Settings{
        URL:         cfg.Backend.Ollama.URL,
        Token:       cfg.Backend.Ollama.Token,
        Model:       curModel,
        ModelList:   modelList,
        CancelFunc:  cancel,
        DialogID:    GenerateID(),
    }
    
    // **Backend Settings**
    history := map[string]interface{}{}
    history["messages"] = []common.LLMMessage{}

    widgets := MainLayout(window, &settings, &history)
    widgets.InputEntry.OnSubmitted = func(text string) {
        history["messages"] = append(history["messages"].([]common.LLMMessage), common.LLMMessage{Role: "User", Content: text})
        widgets.ChatDisplay.SetText(widgets.ChatDisplay.Text + fmt.Sprintf("%s\n%s\n\n", common.CHAT_USER_INFO, text))
        widgets.InputEntry.SetText("")
        widgets.ChatScroll.ScrollToBottom()

        ctx, cancel = context.WithCancel(context.Background())
        settings.CancelFunc = cancel
        go ProcessStream(ctx, settings, widgets, history)
    }

    // **APP Start**
    window.SetContent(widgets.MainSplit)
    window.ShowAndRun()
}



