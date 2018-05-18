package main

import (
	"tetra/lib/gui"
	"tetra/lib/store"
)

// Window wrap operating systems's window object.
type Window struct {
	gui.Window // super
}

// OnCreate event handler
func (w *Window) OnCreate() {
	w.Window.OnCreate()
	id := w.ObjID()
	if id != "" {
		if err := store.LoadState("state", id, w); err != nil {
			store.LoadState("layout", "default", w)
		}
	}
}

// OnDestroy event handler
func (w *Window) OnDestroy() {
	id := w.ObjID()
	if id != "" {
		store.SaveState("state", id, w)
	}
	w.Window.OnDestroy()
}
