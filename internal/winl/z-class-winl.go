package winl

// Auto generated file, do NOT edit!

import "tetra/lib/factory"

var factoryRegisted bool

// FactoryRegister register creator in factory for package winl
func FactoryRegister() {
	if factoryRegisted {
		return
	}
	factoryRegisted = true

	factory.Register(`winl.Window`, func() interface{} {
		return NewWindow()
	})
}

// NewWindow create and init new Window object.
func NewWindow() *Window {
	p := new(Window)
	p.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Window) Class() string {
	return (`winl.Window`)
}

// IWindow is interface of class Window
type IWindow interface {
	// Class name for factory
	Class() string
	// Create the window, width and height can be zero.
	Create(width, height int) error
	// Destroy the window
	Destroy()
	// Expose triggle expose event
	Expose(x, y, width, height float32)
	// Init the object
	Init()
	// IsFullScreen determine if window is full screen
	IsFullScreen() bool
	// IsVisible determine if window is visible
	IsVisible() bool
	// MakeCurrent set current OpenGL to this window
	MakeCurrent() bool
	// OnCreate event handler
	OnCreate()
	// OnDestroy event handler
	OnDestroy()
	// OnExpose event handler
	OnExpose(x, y, width, height float32)
	// OnMouseEnter event handler
	OnMouseEnter(x, y float32)
	// OnMouseLeave event handler
	OnMouseLeave(x, y float32)
	// OnMouseMove event handler
	OnMouseMove(x, y float32)
	// OnMousePress event handler
	OnMousePress(btn int, x, y float32)
	// OnMouseRelease event handler
	OnMouseRelease(btn int, x, y float32)
	// OnMouseWheel event handler
	OnMouseWheel(vert bool, dz float32)
	// OnResize event handler
	OnResize(width, height float32)
	// Present copy OpenGL content from back buffer to front buffer, make it visible
	Present()
	// SetHints set hints for window style
	SetHints(hints hints)
	// SetTitle set the window title
	SetTitle(title string)
	// Show the window
	Show()
	// Size reports size of window's client area
	Size() (width, height float32)
	// ToggleFullScreen switch between full screen mode and normal mode
	ToggleFullScreen()
}
