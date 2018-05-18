package skin

// Auto generated file, do NOT edit!

import "tetra/lib/factory"

var factoryRegisted bool

// FactoryRegister register creator in factory for package skin
func FactoryRegister() {
	if factoryRegisted {
		return
	}
	factoryRegisted = true

	factory.Register(`skin.Common`, func() interface{} {
		return NewCommon()
	})
	factory.Register(`skin.Fallback`, func() interface{} {
		return NewFallback()
	})
}

// NewCommon create and init new Common object.
func NewCommon() *Common {
	p := new(Common)
	p.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Common) Class() string {
	return (`skin.Common`)
}

// ICommon is interface of class Common
type ICommon interface {
	// Class name for factory
	Class() string
	// Init the object
	Init()
}

// NewFallback create and init new Fallback object.
func NewFallback() *Fallback {
	p := new(Fallback)
	p.Common.Self = p
	p.Init()
	return p
}

// Class name for factory
func (p *Fallback) Class() string {
	return (`skin.Fallback`)
}

// IFallback is interface of class Fallback
type IFallback interface {
	ICommon
}
