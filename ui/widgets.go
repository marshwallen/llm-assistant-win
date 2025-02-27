package ui

import (
	"container/list"
	"fmt"
	"winds-assistant/common"
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
func MainWidgets(window fyne.Window, settings *common.Settings, history *list.List) (common.Widgets){
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
        widget.NewButton(common.WIDGET_SWITCH_AGENT, func() {
            if settings.EnableAgent{
                settings.SysPrompt = workers.SYSTEM_PROMPT_DEFAULT
                settings.EnableAgent = false
            }else{
                settings.SysPrompt = workers.SYSTEM_PROMPT_WITH_TOOLS
                settings.EnableAgent = true
            }
            updateSidebarInfo(modelTitle, settings)
        }),
        widget.NewButton(common.WIDGET_SETTING, func() {
            showSettingsDialog(window, modelTitle, settings)
        }),
        widget.NewButton(common.WIDGET_REFRESH, func() {
            modelList := workers.GetModelList(settings.URL, window)
            if len(modelList) > 0 {
                settings.Model = modelList[0]
            }
            settings.ModelList = modelList
            updateSidebarInfo(modelTitle, settings)
        }),
        modelTitle,
        layout.NewSpacer(),
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
func showSettingsDialog(parent fyne.Window, modelTitle *widget.Label, settings *common.Settings) {
    url := widget.NewEntry()
    url.SetText(settings.URL)

    apikey := widget.NewEntry()
    apikey.SetText(settings.API_KEY)

    modelSelect := widget.NewSelect([]string{common.WIDGET_LOADING}, func(s string) {})
    modelSelect.SetOptions(settings.ModelList)
    modelSelect.SetSelected(settings.Model)
    
    // 构建对话框内容
    form := widget.NewForm(
        widget.NewFormItem(common.WIDGET_FORM_URL, url),
        widget.NewFormItem(common.WIDGET_FORM_APIKEY, apikey),
        widget.NewFormItem(common.WIDGET_FORM_MODEL_SELECT, modelSelect),
    )

    dialog.ShowCustomConfirm("", common.WIDGET_DIALOG_SAVE, common.WIDGET_DIALOG_CANCEL,
        container.NewVBox(form),
        func(save bool) {
            if save {
                // 获取选中的值
                settings.URL = url.Text
                settings.API_KEY = apikey.Text
                settings.Model = modelSelect.Selected
                updateSidebarInfo(modelTitle, settings)
            }
        }, parent)
}

// 更新侧边栏信息
func updateSidebarInfo(sidebar *widget.Label, settings *common.Settings) {
    sidebar.SetText(fmt.Sprintf("%s\n%s\n\n%s\n%s\n\n%s\n%s\n\n%s\n%v\n", 
        common.SYSTEM_URL_INFO,
        settings.URL,
        common.SYSTEM_MODEL_INFO,
        settings.Model,
        common.SYSTEM_DIALOG_ID_INFO,
        settings.DialogID,
        common.SYSTEM_AGENT_STATUS_INFO,
        settings.EnableAgent,
    ))
}