package gui

// Pane is compound widget typically use as split area in window.
type Pane struct {
	Widget
}

// State to string
func (pn *Pane) State() ([]byte, error) {
	return nil, nil
}

// SetState from string
func (pn *Pane) SetState(data []byte) error {
	return nil
}

// Is3D reports whether pane is 3D scene
func (pn *Pane) Is3D() bool {
	return false
}
