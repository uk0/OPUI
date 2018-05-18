package gui

// Auto generated file, do NOT edit!

import (
	"tetra/internal/winl"
	"tetra/lib/factory"
	"tetra/lib/glman"
)

var factoryRegisted bool

// FactoryRegister register creator in factory for package gui
func FactoryRegister() {
	if factoryRegisted {
		return
	}
	factoryRegisted = true

	factory.Register(`gui.Button`, func() interface{} {
		return NewButton()
	})
	factory.Register(`gui.Elem`, func() interface{} {
		return NewElem()
	})
	factory.Register(`gui.Pane`, func() interface{} {
		return NewPane()
	})
	factory.Register(`gui.Pane3D`, func() interface{} {
		return NewPane3D()
	})
	factory.Register(`gui.TestPane`, func() interface{} {
		return NewTestPane()
	})
	factory.Register(`gui.TestPane3D`, func() interface{} {
		return NewTestPane3D()
	})
	factory.Register(`gui.Widget`, func() interface{} {
		return NewWidget()
	})
	factory.Register(`gui.Window`, func() interface{} {
		return NewWindow()
	})
}

// NewButton create and init new Button object.
func NewButton() *Button {
	p := new(Button)
	p.Widget.Elem.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Button) Class() string {
	return (`gui.Button`)
}

// IButton is interface of class Button
type IButton interface {
	IWidget
	// Font returns current font
	Font() glman.Font
	// SetFont set the font
	SetFont(f glman.Font)
	// SetText set the text label on the button
	SetText(s string)
	// Text label on the button
	Text() string
}

// NewElem create and init new Elem object.
func NewElem() *Elem {
	p := new(Elem)
	p.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Elem) Class() string {
	return (`gui.Elem`)
}

// IElem is interface of class Elem
type IElem interface {
	// Bounds reports bounds rect of the element
	Bounds() Rect
	// BoundsGLCoord reports bounds rect of the element, in OpenGL (Y-UP) coordinate.
	BoundsGLCoord() Rect
	// Children returns child elements
	Children() []IElem
	// Class name for factory
	Class() string
	// Index of x
	Index(x IElem) int
	// Init a new object
	Init()
	// Insert x at index i, if i < 0 then append to the end
	Insert(i int, x IElem)
	// Parent returns parent element
	Parent() IElem
	// Remove child at index i
	Remove(i int) IElem
	// RemoveAll remove all children
	RemoveAll()
	// Render the element
	Render()
	// SetBounds set the bounds rect of the element
	SetBounds(rect Rect)
	// SetParent set parent element
	SetParent(p IElem)
	// SetWindow set the owner window
	SetWindow(w IWindow)
	// Window reports the owner window
	Window() IWindow
}

// NewPane create and init new Pane object.
func NewPane() *Pane {
	p := new(Pane)
	p.Widget.Elem.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Pane) Class() string {
	return (`gui.Pane`)
}

// IPane is interface of class Pane
type IPane interface {
	IWidget
	// Is3D reports whether pane is 3D scene
	Is3D() bool
	// SetState from string
	SetState(data []byte) error
	// State to string
	State() ([]byte, error)
}

// NewPane3D create and init new Pane3D object.
func NewPane3D() *Pane3D {
	p := new(Pane3D)
	p.Pane.Widget.Elem.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Pane3D) Class() string {
	return (`gui.Pane3D`)
}

// IPane3D is interface of class Pane3D
type IPane3D interface {
	IPane
	// Frustum set projection to special frustum
	Frustum(left float32, right float32, bottom float32, top float32, near float32, far float32)
	// LookAt set view matrix to look at special point
	LookAt(eye Vec3, lookAt Vec3, up Vec3)
	// Ortho set projection to orthographic
	Ortho(xMin float32, xMax float32, yMin float32, yMax float32, zMin float32, zMax float32)
	// Perspective set projection to perspective
	Perspective(fovy float32, aspect float32, near float32, far float32)
}

// NewTestPane create and init new TestPane object.
func NewTestPane() *TestPane {
	p := new(TestPane)
	p.Pane.Widget.Elem.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *TestPane) Class() string {
	return (`gui.TestPane`)
}

// ITestPane is interface of class TestPane
type ITestPane interface {
	IPane
}

// NewTestPane3D create and init new TestPane3D object.
func NewTestPane3D() *TestPane3D {
	p := new(TestPane3D)
	p.Pane3D.Pane.Widget.Elem.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *TestPane3D) Class() string {
	return (`gui.TestPane3D`)
}

// ITestPane3D is interface of class TestPane3D
type ITestPane3D interface {
	IPane3D
}

// NewWidget create and init new Widget object.
func NewWidget() *Widget {
	p := new(Widget)
	p.Elem.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Widget) Class() string {
	return (`gui.Widget`)
}

// IWidget is interface of class Widget
type IWidget interface {
	IElem
}

// NewWindow create and init new Window object.
func NewWindow() *Window {
	p := new(Window)
	p.Window.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Window) Class() string {
	return (`gui.Window`)
}

// IWindow is interface of class Window
type IWindow interface {
	winl.IWindow
	// Layout return current split layout
	Layout() *WndLayout
	// ObjID returns the object id
	ObjID() string
	// OnSkin handle the skin change event
	OnSkin()
	// Render the scene
	Render()
	// SetLayout set the split layout
	SetLayout(wl *WndLayout) error
	// SetObjID set the object id
	SetObjID(id string)
	// SetState from string
	SetState(data []byte) error
	// State to string
	State() ([]byte, error)
}
