package ui

import (
	"container/list"
	"fmt"
	"strings"
	"winds-assistant/common"
	"winds-assistant/utils"
	"winds-assistant/workers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

const (
    // 输入框的最大长度
    MaxInputLength = 2048
    // 聊天窗口一次渲染的最大文本长度
    MaxChatMessageLength = 2048
    // 聊天滑动窗口单边缓冲区大小
    SideCacheSize = 64   
)

// 创建并返回一个主布局，包含侧边栏和聊天窗口
func MainWidgets(window fyne.Window, history *list.List, settings *common.Settings, cfg *common.LLMConfig) (common.Widgets){
    // 侧边信息栏
    modelTitle := widget.NewLabel("") 
    modelTitle.Wrapping = fyne.TextWrapWord
    updateSidebarInfo(modelTitle, settings)
    
    // 聊天窗口显示
    chatDisplay := widget.NewLabel(common.SYSTEM_CHAT_INFO)
    chatDisplay.Wrapping = fyne.TextWrapWord

    // 保存分块的聊天信息
    chatChunk := &common.ChatChunkProcessor{WindowSize: MaxChatMessageLength, SideCacheSize: SideCacheSize}

    // 输入框
    inputEntry := widget.NewMultiLineEntry()
    inputEntry.SetPlaceHolder(common.WIDGET_INPUT_PLACEHOLDER)
    inputEntry.Wrapping = fyne.TextWrapWord

    // 动态截断超长输入
    inputEntry.OnChanged = func(text string) {
        if len(text) > MaxInputLength {
            // 立即截断到允许长度，避免卡死
            trimmed := text[:MaxInputLength]
            inputEntry.SetText(trimmed)  // 自动触发刷新
        }
    }
    
    // 聊天窗口滚动
    chatScroll := common.NewSmartScroll(chatDisplay)

    chatScroll.OnScrollToTop = func() {
        chatDisplay.SetText(chatChunk.RenderPreText())
    }

    chatScroll.OnScrollToBottom = func() {
        chatDisplay.SetText(chatChunk.RenderNextText())
    }

    // 聊天窗口布局
    chatSplit := container.NewVSplit(
        chatScroll,
        container.NewBorder(nil, nil, nil, 
            widget.NewButton(common.WIDGET_SEND, func() {
                if inputEntry.Text != "" {
                    inputEntry.OnSubmitted(inputEntry.Text)
                }
            }),
            inputEntry,
        ),
    )
    chatSplit.SetOffset(0.75)

    // 侧边栏组件
    sidebar := container.NewVBox(
        widget.NewButton(common.WIDGET_NEW_CHAT, func() {
            if settings.CancelFunc != nil {
                settings.CancelFunc()
            }
            history.Init()

            settings.DialogID = GenerateID()
            updateSidebarInfo(modelTitle, settings)

            chatChunk.ClearChunks()
            chatChunk.Process(common.SYSTEM_CHAT_INFO)
        	chatDisplay.SetText(chatChunk.RenderNextText())
            chatDisplay.Refresh()
        }),
        widget.NewButton(common.WIDGET_CHAT_TERMINATE, func() {
            if settings.CancelFunc != nil {
                settings.CancelFunc()
            }
            settings.Running = false
        }),
        widget.NewButton(common.WIDGET_COPY_CHAT, func() {
            clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
            clipboard.SetContent(chatDisplay.Text)
        }),
        widget.NewButton(common.WIDGET_AGENT_SWITCH, func() {
            if settings.EnableAgent{
                settings.SysPrompt = workers.SYSTEM_PROMPT_DEFAULT
                settings.EnableAgent = false
            }else{
                settings.SysPrompt = workers.SYSTEM_PROMPT_WITH_TOOLS_BASE + strings.Join(workers.ToolsPromptRegister, "\n")
                settings.EnableAgent = true
            }
            updateSidebarInfo(modelTitle, settings)
        }),
        widget.NewButton(common.WIDGET_SETTING, func() {
            if settings.Running{
                common.ShowErrorDialog(window, fmt.Errorf("info: assistant is running, terminate it first"))
                return
            }
            showSettingsDialog(window, modelTitle, settings, cfg)
        }),
        widget.NewButton(common.WIDGET_REFRESH, func() {
            if settings.BackendName == "ollama" {
                modelList := workers.GetModelList(settings.BackendCfg.BaseURL, window)
                if len(modelList) > 0 {
                    settings.BackendCfg.Model = modelList[0]
                }
                settings.ModelList = modelList
            }
            updateSidebarInfo(modelTitle, settings)
        }),
        modelTitle,
        layout.NewSpacer(),
        widget.NewButton(common.WIDGET_FASTCLIBOARD, func() {
            showFastCliboard(window, settings)
        }),
        widget.NewButton(common.WIDGET_BACKEND_SETTING, func() {
            if settings.Running{
                common.ShowErrorDialog(window, fmt.Errorf("info: assistant is running, terminate it first"))
                return
            }
            showBackendSettingDialog(window, modelTitle, settings, cfg)
        }),
    )

    // 主布局
    mainSplit := container.NewHSplit(
        sidebar,
        chatSplit,
    )
    mainSplit.SetOffset(0.2)
	return common.Widgets{
        Window:         window,
		MainSplit: 		mainSplit,
		ChatDisplay: 	chatDisplay,
		ChatScroll: 	chatScroll,
		InputEntry: 	inputEntry,
        ChatChunk:      chatChunk,
	}
}

// 设置对话框
func showSettingsDialog(parent fyne.Window, modelTitle *widget.Label, settings *common.Settings, cfg *common.LLMConfig) {
    url := widget.NewEntry()
    url.SetText(settings.BackendCfg.BaseURL)

    apikey := widget.NewEntry()
    apikey.SetText(settings.BackendCfg.APIKey)

    modelSelect := widget.NewSelect([]string{common.WIDGET_LOADING}, func(s string) {})
    modelSelect.SetOptions(settings.ModelList)
    modelSelect.SetSelected(settings.BackendCfg.Model)

    model := widget.NewEntry()
    model.SetText(settings.BackendCfg.Model)

    // 只有ollama后端可以自主选择模型
    if settings.BackendName == "ollama" {
        model.Disable()
    }else{
        modelSelect.Disable()
        url.Disable()
    }
    
    // 构建对话框内容
    form := widget.NewForm(
        widget.NewFormItem(common.WIDGET_FORM_URL, url),
        widget.NewFormItem(common.WIDGET_FORM_APIKEY, apikey),
        widget.NewFormItem(common.WIDGET_FORM_MODEL_SELECT, modelSelect),
        widget.NewFormItem(common.WIDGET_CURRENT_MODEL, model),
    )

    dialog.ShowCustomConfirm("", common.WIDGET_DIALOG_SAVE, common.WIDGET_DIALOG_CANCEL,
        container.NewVBox(form),
        func(save bool) {
            if save {
                // 获取选中的值
                settings.BackendCfg.BaseURL = url.Text
                settings.BackendCfg.APIKey = apikey.Text
                if settings.BackendName == "ollama" {
                    settings.BackendCfg.Model = modelSelect.Selected
                    model.SetText(settings.BackendCfg.Model)
                }else{
                    settings.BackendCfg.Model = model.Text 
                    settings.ModelList = []string{settings.BackendCfg.Model}
                }
                updateSidebarInfo(modelTitle, settings)

                // 保存配置到文件中
                cfg.Backend[settings.BackendName] = settings.BackendCfg
                cfg.Default = settings.BackendName
                utils.SaveCfg(cfg)
            }
        }, parent)
}

// 更新侧边栏信息
func updateSidebarInfo(sidebar *widget.Label, settings *common.Settings) {
    sideText := []string{
        // Backend
        common.SYSTEM_BACKEND_INFO,
        common.BACKEND_MAP[settings.BackendName],
        "",

        // URL
        common.SYSTEM_URL_INFO,
        settings.BackendCfg.BaseURL,
        "",

        // Model
        common.SYSTEM_MODEL_INFO,
        settings.BackendCfg.Model,
        "",

        // Dialog ID
        common.SYSTEM_DIALOG_ID_INFO,
        settings.DialogID,
        "",

        // Agent Status
        common.SYSTEM_AGENT_STATUS_INFO,
        fmt.Sprintf("%v", settings.EnableAgent),
    }
    sidebar.SetText(strings.Join(sideText, "\n"))
}

func showBackendSettingDialog(parent fyne.Window, modelTitle *widget.Label, settings *common.Settings, cfg *common.LLMConfig) {
    // Backend 选择器
    backendSelect := widget.NewSelect([]string{common.WIDGET_LOADING}, func(s string) {})
    var backends []string
    for name := range cfg.Backend {
        backends = append(backends, name)
    }
    backendSelect.SetOptions(backends)
    backendSelect.SetSelected(settings.BackendName)

    // 构建对话框内容
    form := widget.NewForm(
        widget.NewFormItem(common.WIDGET_FORM_BACKEND, backendSelect),
    )

    dialog.ShowCustomConfirm("", common.WIDGET_DIALOG_SAVE, common.WIDGET_DIALOG_CANCEL,
        container.NewVBox(form),
        func(save bool) {
            if save {
                choice := backendSelect.Selected
                settings.BackendName = choice
                settings.BackendCfg = cfg.Backend[choice]
                if choice == "ollama" {
                    settings.ModelList = workers.GetModelList(settings.BackendCfg.BaseURL, parent)
                }

                // 保存配置文件
                cfg.Default = choice
                utils.SaveCfg(cfg)
                // 刷新
                updateSidebarInfo(modelTitle, settings)
            }
        }, parent)
}

func showFastCliboard(parent fyne.Window, settings *common.Settings) {
    // 快捷指令板
    fastEntry := widget.NewMultiLineEntry()
    fastEntry.Wrapping = fyne.TextWrapWord
    fastEntry.SetText(settings.FastCliboard)
    fastEntry.SetMinRowsVisible(10)

    container := container.NewVBox(fastEntry)
    fastEntry.OnChanged = func(text string) {
        settings.FastCliboard = text
    }

    container.Show()
    dialog.ShowCustomConfirm(
        common.WIDGET_FASTCLIBOARD, 
        "Confirm", "Cancel",
        container, func(save bool) {
            settings.FastCliboard = fastEntry.Text
            utils.EnsureDir("data/")
            utils.WriteTxtFile("data/fast_cliboard.txt", settings.FastCliboard)
        }, parent)
}