package glman

import (
	"errors"
	"tetra/internal/winl"
)

var infClipRect = Rect{-99999, -99999, 99999, 99999}

// OSVersion return the version of operating system
func OSVersion() (s string) {
	return winl.OSVersion()
}

// Clip2DStack stack type for 2D clipping
type Clip2DStack []Rect

// Peek stack top
func (s *Clip2DStack) Peek() Rect {
	return (*s)[len(*s)-1]
}

// Load rect into stack top
func (s *Clip2DStack) Load(rect Rect) {
	(*s)[len(*s)-1] = rect
}

// LoadInf load a very large rect, make it act as no clip at all.
// the components is large numbers, not realy INF.
func (s *Clip2DStack) LoadInf(rect Rect) {
	(*s)[len(*s)-1] = infClipRect
}

// Push stack, new top is a very large rect.
func (s *Clip2DStack) Push() {
	*s = append(*s, infClipRect)
}

// Pop stack
func (s *Clip2DStack) Pop() error {
	if len(*s) == 1 {
		return errors.New("Can not pop from clip 2D stack")
	}
	*s = (*s)[:len(*s)-1]
	return nil
}
