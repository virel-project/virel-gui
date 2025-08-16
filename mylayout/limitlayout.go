package mylayout

import (
	"fyne.io/fyne/v2"
)

// Declare conformity with Layout interface
var _ fyne.Layout = (*LimitLayout)(nil)

type LimitLayout struct {
	maxWidth  float32
	maxHeight float32
}

// NewLimitLayout creates a new LimitLayout instance
func NewLimitLayout(maxWidth, maxHeight float32) fyne.Layout {
	return &LimitLayout{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

// Layout is called to pack all child objects into a specified size.
// For CenterLayout this sets all children to their minimum size, centered within the space.
func (l *LimitLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	width := min(size.Width, l.maxWidth)
	height := min(size.Height, l.maxHeight)
	for _, child := range objects {
		childMin := fyne.NewSize(width, height)
		child.Resize(childMin)
		child.Move(fyne.NewPos(float32(size.Width-childMin.Width)/2, float32(size.Height-childMin.Height)/2))
	}
}

// MinSize finds the smallest size that satisfies all the child objects.
// For CenterLayout this is determined simply as the MinSize of the largest child.
func (l *LimitLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}

		minSize = minSize.Max(child.MinSize())
	}

	return minSize
}
