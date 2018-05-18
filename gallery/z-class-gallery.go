package main

// Auto generated file, do NOT edit!

import (
	"tetra/lib/factory"
	"tetra/lib/gui"
)

var factoryRegisted bool

// FactoryRegister register creator in factory for package main
func FactoryRegister() {
	if factoryRegisted {
		return
	}
	factoryRegisted = true

	factory.Register(`main.Window`, func() interface{} {
		return NewWindow()
	})
}

// NewWindow create and init new Window object.
func NewWindow() *Window {
	p := new(Window)
	p.Window.Window.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Window) Class() string {
	return (`main.Window`)
}

// IWindow is interface of class Window
type IWindow interface {
	gui.IWindow
}
