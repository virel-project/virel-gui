package mywidget

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Bar is a container that lays out items horizontally with a background color
type Bar struct {
	widget.BaseWidget
	background *canvas.Rectangle
	container  *fyne.Container
}

// NewBar creates a new Bar widget with the specified background color
func NewBar(bgColor color.Color, objects ...fyne.CanvasObject) *Bar {
	b := &Bar{
		background: canvas.NewRectangle(bgColor),
	}
	b.ExtendBaseWidget(b)

	b.container = container.NewHBox(objects...)
	return b
}

// CreateRenderer is required for the widget implementation
func (b *Bar) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(
		container.NewStack(
			b.background,
			b.container,
		),
	)
}

// Add appends the given objects to the bar's content
func (b *Bar) Add(objects ...fyne.CanvasObject) {
	for _, v := range objects {
		b.container.Add(v)
	}
	b.Refresh()
}

// Remove removes the given objects from the bar's content
func (b *Bar) Remove(objects ...fyne.CanvasObject) {
	for _, v := range objects {
		b.container.Remove(v)
	}
	b.Refresh()
}

// SetBackgroundColor changes the bar's background color
func (b *Bar) SetBackgroundColor(c color.Color) {
	b.background.FillColor = c
	b.background.Refresh()
}
