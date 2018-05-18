package gui

// TestPane3D is a pane for 3D scene
type TestPane3D struct {
	Pane3D
}

// Init a new object
func (pn *TestPane3D) Init() {
	pn.Pane3D.Init()
}

// State to string
func (pn *TestPane3D) State() ([]byte, error) {
	//return []byte(pn.btn.Text()), nil
	return nil, nil
}

// SetState from string
func (pn *TestPane3D) SetState(data []byte) error {
	//pn.btn.SetText(string(data))
	return nil
}
