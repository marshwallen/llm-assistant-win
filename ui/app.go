package ui

import (
	"context"
	"fmt"
	"sync"
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
    modelList := workers.GetModelList(cfg.Backend.Ollama.URL, window)

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
        EnableAgent: true,
    }
    
    // **Backend Settings**
    history := map[string]interface{}{}
    history["messages"] = []common.LLMMessage{}
    history["messages"] = append(history["messages"].([]common.LLMMessage), common.LLMMessage{Role: "System", Content: workers.SYSTEM_PROMPT})
    
    widgets := MainWidgets(window, &settings, &history)
    widgets.InputEntry.OnSubmitted = func(text string) {
        // 提交前先检查模型存不存在
        if settings.Model == "" {
            common.ShowErrorDialog(window, fmt.Errorf("error: model not found"))
            return
        }
        history["messages"] = append(history["messages"].([]common.LLMMessage), common.LLMMessage{Role: "User", Content: text})
        widgets.ChatDisplay.SetText(widgets.ChatDisplay.Text + fmt.Sprintf("%s\n%s\n\n", common.CHAT_USER_INFO, text))
        widgets.InputEntry.SetText("")
        widgets.ChatScroll.ScrollToBottom()

        ctx, cancel = context.WithCancel(context.Background())
        settings.CancelFunc = cancel

        var wg sync.WaitGroup
        wg.Add(1)
        go ProcessStream(ctx, &wg, settings, widgets, history)
        if settings.EnableAgent{
            wg.Wait()
            messagesList := history["messages"].([]common.LLMMessage)
            lastMessage := messagesList[len(messagesList)-1]
            if lastMessage.Role == "Agent" {
                history["messages"] = append(history["messages"].([]common.LLMMessage), common.LLMMessage{
                    Role: "User", Content: workers.USER_PROMPT_LAST + lastMessage.Content,
                })
                wg.Add(1)
                go ProcessStream(ctx, &wg, settings, widgets, history)
            }  
        }
        wg.Wait()
    }

    // **APP Start**
    window.SetContent(widgets.MainSplit)
    window.ShowAndRun()
}



