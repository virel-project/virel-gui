package mywidget

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Card struct {
	widget.BaseWidget

	Title      string
	Comment    string
	CopiedText string

	app fyne.App

	bgColor color.Color

	titleObj   *canvas.Text
	commentObj *canvas.Text

	container *fyne.Container

	lastTapped time.Time
}

func NewCard(app fyne.App, bgColor color.Color, title, comment, copiedtext string) *Card {
	titleObj := canvas.NewText(title, theme.Color(theme.ColorNameForeground))
	titleObj.TextStyle.Bold = true
	titleObj.TextSize = theme.TextSubHeadingSize()

	commentObj := canvas.NewText(comment, theme.Color(theme.ColorNameForeground))
	commentObj.Alignment = fyne.TextAlignCenter

	item := &Card{
		Title:      title,
		Comment:    comment,
		CopiedText: copiedtext,
		titleObj:   titleObj,
		commentObj: commentObj,
		app:        app,
		bgColor:    bgColor,
	}
	item.ExtendBaseWidget(item)

	return item
}

func (c *Card) SetTitle(title string) {
	c.Title = title
	c.titleObj.Text = title
	c.titleObj.Refresh()
}
func (c *Card) SetComment(comment string) {
	c.Comment = comment
	c.commentObj.Text = comment
	c.commentObj.Refresh()
}

func (item *Card) CreateRenderer() fyne.WidgetRenderer {
	fmt.Println("addressBox create renderer")

	rect := canvas.NewRectangle(item.bgColor)
	rect.CornerRadius = theme.InputRadiusSize()

	icon := widget.NewIcon(theme.ContentCopyIcon())

	/*iconCanvas := canvas.NewImageFromResource(theme.ContentCopyIcon())

	iconCanvas.SetMinSize(fyne.NewSquareSize(theme.IconInlineSize()))*/

	c := container.NewStack(rect, container.NewPadded(container.NewVBox(
		container.NewStack(container.NewCenter(item.titleObj), container.NewBorder(nil, nil, nil, icon)),
		item.commentObj,
	)))

	item.container = c

	return widget.NewSimpleRenderer(c)
}

func (a *Card) Resize(s fyne.Size) {
	txt := a.Title
	reduced := false
	for len(txt) > 5 && s.Width < fyne.MeasureText(txt+"...ICONHERE", theme.TextSubHeadingSize(), fyne.TextStyle{}).Width {
		reduced = true
		txt = txt[:len(txt)-2]
	}
	if reduced {
		txt = txt + "..."
	}
	a.titleObj.Text = txt

	a.BaseWidget.Resize(s)
}

func (a *Card) MinSize() fyne.Size {
	ms := a.BaseWidget.MinSize()

	ms.Width = 150

	return ms
}
func (a *Card) Refresh() {
	a.BaseWidget.Refresh()
	a.titleObj.Refresh()
	a.commentObj.Refresh()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (a *Card) Tapped(ev *fyne.PointEvent) {
	if time.Since(a.lastTapped) < 5*time.Second {
		return
	}

	fmt.Println("tapped!")
	a.lastTapped = time.Now()

	a.app.Clipboard().SetContent(a.Title)

	fyne.Do(func() {
		a.commentObj.Text = a.CopiedText
		a.Refresh()
	})

	go func() {
		time.Sleep(5 * time.Second)
		fyne.Do(func() {
			a.commentObj.Text = a.Comment
			a.Refresh()
		})
	}()

}
