package common

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// 在给定的父窗口中显示错误对话框
func ShowErrorDialog(parent fyne.Window, err error) {
	errLabel := widget.NewLabel(err.Error())
	errLabel.Wrapping = fyne.TextWrapWord
    dialog.ShowCustomConfirm("Error", "Confirm", "Cancel", errLabel, 
		func (confirm bool) {},
		parent,
	)
}