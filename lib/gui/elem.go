package gui

import (
	"tetra/lib/geom"
	"tetra/lib/glman"
	//	"tetra/lib/geom"
	//	"tetra/lib/glman"
)

// Elem class is abstract gui element
type Elem struct {
	Self   interface{}
	bounds Rect
	parent IElem
	child  []IElem
	wnd    IWindow
}

// Init a new object
func (el *Elem) Init() {
}

// Parent returns parent element
func (el *Elem) Parent() IElem {
	return el.parent
}

// SetParent set parent element
func (el *Elem) SetParent(p IElem) {
	el.parent = p
}

// Children returns child elements
func (el *Elem) Children() []IElem {
	return el.child
}

// Bounds reports bounds rect of the element
func (el *Elem) Bounds() Rect {
	return el.bounds
}

// BoundsGLCoord reports bounds rect of the element, in OpenGL (Y-UP) coordinate.
func (el *Elem) BoundsGLCoord() (rc Rect) {
	rc = el.bounds
	win := el.Window()
	_, h := win.Size()
	rc[1], rc[3] = h-rc[3], h-rc[1]
	return rc
}

// SetBounds set the bounds rect of the element
func (el *Elem) SetBounds(rect Rect) {
	el.bounds = rect
}

// Window reports the owner window
func (el *Elem) Window() IWindow {
	if el.wnd != nil {
		return el.wnd
	}
	if el.parent != nil {
		el.wnd = el.parent.Window()
	}
	return el.wnd
}

// SetWindow set the owner window
func (el *Elem) SetWindow(w IWindow) {
	el.wnd = w
	for _, c := range el.child {
		c.SetWindow(w)
	}
}

// Insert x at index i, if i < 0 then append to the end
func (el *Elem) Insert(i int, x IElem) {
	if i < 0 {
		el.child = append(el.child, x)
	} else {
		el.child = append(el.child, nil)
		copy(el.child[i+1:], el.child[i:])
		el.child[i] = x
	}
}

// Remove child at index i
func (el *Elem) Remove(i int) IElem {
	x := el.child[i]
	x.SetParent(nil)
	copy(el.child[i:], el.child[i+1:])
	el.child[len(el.child)-1] = nil
	el.child = el.child[:len(el.child)-1]
	return x
}

// RemoveAll remove all children
func (el *Elem) RemoveAll() {
	for _, c := range el.child {
		c.SetParent(nil)
	}
	el.child = nil
}

// Index of x
func (el *Elem) Index(x IElem) int {
	for i, c := range el.child {
		if c == x {
			return i
		}
	}
	return -1
}

func round(x float32) float32 {
	return float32(int(x + 0.5))
}

// Render the element
func (el *Elem) Render() {
	glman.StackMatM.Push()
	glman.StackMatM.Load(geom.Mat4Trans(round(el.bounds.X0()), round(el.bounds.Y0()), 0))
	glman.StackClip2D.Push()
	rect := el.bounds
	//dbg.Logf("rect=%v\n", rect)
	glman.StackClip2D.Load(rect)
	for _, c := range el.child {
		c.Render()
	}
	glman.StackClip2D.Pop()
	glman.StackMatM.Pop()
}
