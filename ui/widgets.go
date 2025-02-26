package ui

import (
	"fmt"
	"winds-assistant/common"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// 创建并返回一个主布局，包含侧边栏和聊天窗口
func MainLayout(window fyne.Window, settings *common.Settings, history *map[string]interface{}) (common.Widgets){
    
    // 侧边信息栏
    modelTitle := widget.NewLabel("") 
    modelTitle.Wrapping = fyne.TextWrapWord
    updateSidebarInfo(modelTitle, settings)
    
    // 聊天窗口显示
    chatDisplay := widget.NewLabel(common.SYSTEM_CHAT_INFO)
    chatDisplay.Wrapping = fyne.TextWrapWord

    inputEntry := widget.NewEntry()
    inputEntry.SetPlaceHolder(common.WIDGET_INPUT_PLACEHOLDER)
    
    // 聊天窗口布局
    chatScroll := container.NewScroll(chatDisplay)
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
            *history = make(map[string]interface{})
            (*history)["messages"] = []common.LLMMessage{}

            settings.DialogID = GenerateID()
            updateSidebarInfo(modelTitle, settings)
            chatDisplay.SetText(common.SYSTEM_CHAT_INFO)
        }),
        widget.NewButton(common.WIDGET_STOP, func() {
            if settings.CancelFunc != nil {
                settings.CancelFunc()
            }
        }),
        widget.NewButton(common.WIDGET_SETTING, func() {
            showSettingsDialog(window, modelTitle, settings)
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
		MainSplit: 		mainSplit,
		ChatDisplay: 	chatDisplay,
		ChatScroll: 	chatScroll,
		InputEntry: 	inputEntry,
	}
}

// 设置对话框
func showSettingsDialog(parent fyne.Window, modelTitle *widget.Label, settings *common.Settings) {
    apikey := widget.NewEntry()
    modelSelect := widget.NewSelect([]string{common.WIDGET_LOADING}, func(s string) {})
    modelSelect.SetOptions(settings.ModelList)
    modelSelect.SetSelected(settings.Model)
    apikey.SetText(settings.Token)

    // 构建对话框内容
    form := widget.NewForm(
        widget.NewFormItem(common.WIDGET_FORM_APIKEY, apikey),
        widget.NewFormItem(common.WIDGET_FORM_MODEL_SELECT, modelSelect),
    )

    dialog.ShowCustomConfirm("", common.WIDGET_DIALOG_SAVE, common.WIDGET_DIALOG_CANCEL,
        container.NewVBox(form),
        func(save bool) {
            if save {
                // 获取选中的值
                settings.Token = apikey.Text
                settings.Model = modelSelect.Selected
                updateSidebarInfo(modelTitle, settings)
            }
        }, parent)
}

// 更新侧边栏信息
func updateSidebarInfo(sidebar *widget.Label, settings *common.Settings) {
    sidebar.SetText(fmt.Sprintf("%s\n%s\n\n%s\n%s\n\n%s\n%s\n", 
        common.SYSTEM_URL_INFO,
        settings.URL,
        common.SYSTEM_MODEL_INFO,
        settings.Model,
        common.SYSTEM_DIALOG_ID_INFO,
        settings.DialogID,
    ))
}