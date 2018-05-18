// Package refc is implement simple reference counting
package refc

import (
	"log"
	"sync/atomic"
)

// Interface for reference counting object
type Interface interface {
	RetainCount() int32
	Retain()
	Release()
}

// Obj is for embed into struct
type Obj struct {
	rc    int32  // retain count
	final func() // finalizer function
}

// Retain the object, increase retain count by 1
func (o *Obj) Retain() {
	newrc := atomic.AddInt32(&o.rc, 1)
	if newrc <= 0 {
		log.Panic("try retain a finalized object")
	}
}

// RetainCount return the retain count
func (o *Obj) RetainCount() int32 {
	return atomic.LoadInt32(&o.rc)
}

// Release the reference, decrease retain count by 1
func (o *Obj) Release() {
	newrc := atomic.AddInt32(&o.rc, -1)
	if newrc > 0 {
		return
	}
	if newrc == 0 {
		atomic.AddInt32(&o.rc, -1) // trick mark object finalized
		o.final()
		return
	}
	log.Panic("over-release a object")
}

// SetFinalizer set finalizer for object
func SetFinalizer(o *Obj, fn func()) {
	o.final = fn
}
