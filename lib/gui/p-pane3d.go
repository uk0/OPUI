package gui

import (
	"tetra/lib/geom"
	"tetra/lib/glman"
)

// Pane3D is a pane for 3D scene
type Pane3D struct {
	Pane
	MatP geom.Mat4
	MatV geom.Mat4
}

// Init a new object
func (pn *Pane3D) Init() {
	pn.Pane.Init()
	pn.MatP = geom.Mat4Ident()
	pn.MatV = geom.Mat4Ident()
}

// State to string
func (pn *Pane3D) State() ([]byte, error) {
	//return []byte(pn.btn.Text()), nil
	return nil, nil
}

// SetState from string
func (pn *Pane3D) SetState(data []byte) error {
	//pn.btn.SetText(string(data))
	return nil
}

// Is3D reports whether pane is 3D scene
func (pn *Pane3D) Is3D() bool {
	return true
}

// Ortho set projection to orthographic
func (pn *Pane3D) Ortho(xMin float32, xMax float32, yMin float32, yMax float32, zMin float32, zMax float32) {
	pn.MatP = geom.Mat4Ortho(xMin, xMax, yMin, yMax, zMin, zMax)
}

// Frustum set projection to special frustum
func (pn *Pane3D) Frustum(left float32, right float32, bottom float32, top float32, near float32, far float32) {
	pn.MatP = geom.Mat4Frustum(left, right, bottom, top, near, far)
}

// Perspective set projection to perspective
func (pn *Pane3D) Perspective(fovy float32, aspect float32, near float32, far float32) {
	pn.MatP = geom.Mat4Perspective(fovy, aspect, near, far)
}

// LookAt set view matrix to look at special point
func (pn *Pane3D) LookAt(eye Vec3, lookAt Vec3, up Vec3) {
	pn.MatV = geom.Mat4LookAt(eye, lookAt, up)
}

// Render the pane
func (pn *Pane3D) Render() {
	viewportBak := glman.GetViewport()
	//fmt.Println(viewportBak)
	glman.SetViewport(pn.BoundsGLCoord())
	glman.StackMatP.Push()
	glman.StackMatV.Push()
	glman.StackMatM.Push()
	glman.StackMatP.Load(pn.MatP)
	glman.StackMatV.Load(pn.MatV)
	glman.StackMatM.Load(geom.Mat4Ident())
	glman.StackClip2D.Push()
	rect := pn.bounds
	glman.StackClip2D.Load(rect)

	for _, c := range pn.child {
		c.Render()
	}
	glman.StackClip2D.Pop()
	glman.StackMatM.Pop()
	glman.StackMatV.Pop()
	glman.StackMatP.Pop()
	glman.SetViewport(viewportBak)
}
