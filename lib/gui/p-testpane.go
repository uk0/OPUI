package gui

import "tetra/lib/glman"

// TestPane is a pane for testing
type TestPane struct {
	Pane
	btn IButton
}

// Init a new object
func (pn *TestPane) Init() {
	pn.btn = NewButton()
	pn.btn.SetFont(glman.LoadFont("WQY-ZenHei", 20))
	pn.Insert(-1, pn.btn)
}

// State to string
func (pn *TestPane) State() ([]byte, error) {
	return []byte(pn.btn.Text()), nil
}

// SetState from string
func (pn *TestPane) SetState(data []byte) error {
	pn.btn.SetText(string(data))
	return nil
}
