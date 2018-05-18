package main

import (
	"log"
	"tetra/lib/app"
	"tetra/lib/gui"
)

//go:generate go run ../cmd/classp/classp.go .

func main() {
	log.Println("Done.")
	app.Run(func() error {
		w1 := NewWindow()
		//w1.Self = w1
		w1.SetObjID("window1")
		w1.SetHints(gui.HintResizable)
		w1.Create(0, 0)
		w1.SetTitle("w1 resizable")
		w1.Show()
		// w1.ToggleFullScreen()
		w1.Destroy()
		w2 := NewWindow()
		w2.SetObjID("window2")
		w2.Create(200, 200)
		w2.SetTitle("w2 fixed")
		w2.Show()
		// w2.ToggleFullScreen()
		// runtime.GC()
		return nil
	}, nil)
}

var qianzi = `
Test。

`
