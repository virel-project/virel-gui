package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
)

func ErrorDialog(w fyne.Window, err error) {
	fmt.Println("ErrorDialog:", err)
	d := dialog.NewError(err, w)
	d.Show()
}
func InfoDialog(w fyne.Window, title string, message string) {
	fmt.Println("InfoDialog:", title, message)
	d := dialog.NewInformation(title, message, w)
	d.Show()
}
func Dialog(w fyne.Window, title string, confirm, cancel string, content fyne.CanvasObject, callback func(bool)) {
	d := dialog.NewCustomConfirm(title, confirm, cancel, content, callback, w)
	d.Show()
}

func NewTitle(t string) *canvas.Text {
	title := canvas.NewText(t, theme.Color(theme.ColorNameForeground))
	title.TextStyle.Bold = true
	title.TextSize = theme.TextSubHeadingSize()
	title.Alignment = fyne.TextAlignCenter

	return title
}
