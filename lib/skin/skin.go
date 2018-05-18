package skin

import (
	"tetra/internal/winl"
)

//go:generate go run ../../cmd/classp/classp.go .

var (
	cur Interface = NewFallback() // the active skin
)

// Interface is skin interface for gui looks
type Interface interface {
	SizeSplit() float32
}

// Get current skin
func Get() Interface {
	return cur
}

// Set current skin
func Set(skin Interface) {
	cur = skin
	for _, w := range winl.List() {
		if i, ok := w.Self.(interface {
			OnSkin()
		}); ok {
			i.OnSkin()
		}
	}
}

// Common data for skin
type Common struct {
	Self    Interface
	SzSplit float32
}

// Init the object
func (c *Common) Init() {
	c.SzSplit = 6
}

// SizeSplit reports size of splitter
func (c Common) SizeSplit() float32 {
	return c.SzSplit
}
