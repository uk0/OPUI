package main

import (
	"tetra/internal/winl"
)

func main() {

	winl.Run(func() error {
		w1 := winl.NewWindow()
		w1.SetHints(winl.HintResizable)
		w1.Create(100, 200)
		w1.SetTitle("FuckSpeed")
		w1.OnMouseEnter(10,20)
		w1.Show()
		return nil
	}, nil)
}
