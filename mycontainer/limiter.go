package mycontainer

import (
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/virel-project/virel-gui/v2/mylayout"
)

func NewWidthLimiter(maxw float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(mylayout.NewLimitLayout(maxw, math.MaxFloat32), objects...)
}
func NewLimiter(maxw float32, maxh float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(mylayout.NewLimitLayout(maxw, maxh), objects...)
}
