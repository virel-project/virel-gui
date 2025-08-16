package mylayout

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*wrapLayout)(nil)

type wrapLayout struct {
	minWidth float32
}

// NewWrapLayout returns a new WrapLayout instance
func NewWrapLayout(minWidth float32) fyne.Layout {
	return &wrapLayout{minWidth}
}

// Layout is called to pack all child objects into a specified size.
// For a WrapLayout this will attempt to lay all the child objects in a row
// and wrap to a new row if the size is not large enough.
func (g *wrapLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	padding := theme.Padding()

	visibleObjects := 0
	// Size taken up by visible objects
	totalWidth := float32(0)
	totalHeight := float32(0)

	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		visibleObjects++
		totalWidth += child.MinSize().Width
		totalHeight += child.MinSize().Height
	}

	if size.Width < g.minWidth { // vBox
		x, y := float32(0), float32(0)
		for _, child := range objects {
			if !child.Visible() {
				continue
			}

			child.Move(fyne.NewPos(x, y))

			height := child.MinSize().Height
			y += padding + height
			child.Resize(fyne.NewSize(size.Width, height))
		}
	} else { // hBox
		x, y := float32(0), float32(0)
		for _, child := range objects {
			if !child.Visible() {
				continue
			}

			child.Move(fyne.NewPos(x, y))

			width := size.Width / float32(visibleObjects) //child.MinSize().Width
			x += padding + width
			child.Resize(fyne.NewSize(width, size.Height))
		}
	}
}

func (g *wrapLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	addPadding := false
	padding := theme.Padding()
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		childMin := child.MinSize()
		minSize.Height = fyne.Max(childMin.Height, minSize.Height)
		minSize.Width += childMin.Width
		if addPadding {
			minSize.Width += padding
		}
		addPadding = true
	}
	return minSize
}
