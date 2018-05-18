package gui

import (
	"runtime"
	"tetra/internal/winl"
	"tetra/lib/dbg"
	"tetra/lib/geom"
	"tetra/lib/glman"
	"tetra/lib/skin"

	"tetra/internal/gl"
)

const (
	// HintResizable is int for window has resize box
	HintResizable = winl.HintResizable

	// HintVideo is hint for window with animated (FPS video) content
	HintVideo = winl.HintVideo

	// HintPainter is hint for 2d drawing using painter algorithm
	HintPainter = winl.HintPainter
)

// Window class wrap operating systems's window object.
type Window struct {

	// Window is underlying winl.Window
	winl.Window // super

	objID string

	layout  *WndLayout
	szSplit float32

	matProj Mat4
	matView Mat4

	vbo uint32
}

// OnSkin handle the skin change event
func (w *Window) OnSkin() {
	dbg.Logf("OnSkin()\n")
	sk := skin.Get()
	w.szSplit = sk.SizeSplit()
}

// OnCreate event handler
func (w *Window) OnCreate() {
	dbg.Logf("OnCreate()\n")
	w.Window.OnCreate()
	w.matView = geom.Mat4Ident()

	tri := [...]float32{
		-100, 0, 0,
		-100, 100, 0,
		100, -100, 0,
		100, 100, 0,
	}

	w.MakeCurrent()
	gl.GenBuffers(1, &w.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, w.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 12*4, gl.Ptr(&tri[0]), gl.STATIC_DRAW)

	gl.Enable(gl.BLEND)
	glman.DbgCheckError()
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	glman.DbgCheckError()
	gl.ShadeModel(gl.SMOOTH) // Enable Smooth Shading
	glman.DbgCheckError()
	gl.Enable(gl.LINE_SMOOTH)
	glman.DbgCheckError()
	gl.Hint(gl.LINE_SMOOTH_HINT, gl.NICEST)
	glman.DbgCheckError()
	gl.Enable(gl.POINT_SMOOTH)
	glman.DbgCheckError()
	gl.Hint(gl.POINT_SMOOTH_HINT, gl.NICEST)
	glman.DbgCheckError()
	gl.Enable(gl.POLYGON_SMOOTH)
	glman.DbgCheckError()
	gl.Hint(gl.POLYGON_SMOOTH_HINT, gl.NICEST)
	glman.DbgCheckError()
	gl.ClearColor(0.7, 0.7, 0.7, 1.0)
	glman.DbgCheckError()
	gl.ClearDepth(1.0)
	glman.DbgCheckError()
}

// OnDestroy event handler
func (w *Window) OnDestroy() {
	dbg.Logf("OnDestroy()\n")
	w.Window.OnDestroy()
}

// OnResize event handler
func (w *Window) OnResize(width, height float32) {
	dbg.Logf("OnResize(%f, %f)\n", width, height)
	w.Window.OnResize(width, height)

	if w.layout != nil {
		w.layout.rc = Rect{0, 0, width, height}
		w.layout.CalcLayout(w.szSplit)
	}
	w.matProj = geom.Mat4Ortho(0, width, height, 0, -1, 1)
	// move origin form center to top-left
	w.matProj = w.matProj.Mult(geom.Mat4Trans(-width/2, -height/2, 0))
	//w.MakeCurrent()
	gl.Viewport(0, 0, int32(width), int32(height))
	if w.IsVisible() {
		//glman.ProgTest()
		w.Render()
	}
}

// OnMouseEnter event handler
func (w *Window) OnMouseEnter(x, y float32) {
	dbg.Logf("OnMouseEnter(%f, %f)\n", x, y)
}

// OnMouseLeave event handler
func (w *Window) OnMouseLeave(x, y float32) {
	dbg.Logf("OnMouseLeave(%f, %f)\n", x, y)
}

// OnMouseMove event handler
func (w *Window) OnMouseMove(x, y float32) {
	//dbg.Logf("OnMouseMove(%f, %f)\n", x, y)
}

// OnMousePress event handler
func (w *Window) OnMousePress(btn int, x, y float32) {
	dbg.Logf("OnMousePress(%d, %f, %f)\n", btn, x, y)
	w.Render()
}

// OnMouseRelease event handler
func (w *Window) OnMouseRelease(btn int, x, y float32) {
	dbg.Logf("OnMouseRelease(%d, %f, %f)\n", btn, x, y)
}

// OnMouseWheel event handler
func (w *Window) OnMouseWheel(vert bool, dz float32) {
	if vert {
		dbg.Logf("OnMouseWheel(vert, %f)\n", dz)
	} else {
		dbg.Logf("OnMouseWheel(horz, %f)\n", dz)
	}
}

// OnExpose event handler
func (w *Window) OnExpose(x, y, width, height float32) {
	dbg.Logf("OnExpose(%g, %g, %g, %g)\n", x, y, width, height)
	w.Render()
}

// ObjID returns the object id
func (w *Window) ObjID() string {
	return w.objID
}

// SetObjID set the object id
func (w *Window) SetObjID(id string) {
	w.objID = id
}

// Layout return current split layout
func (w *Window) Layout() *WndLayout {
	treePrepairSave(w.layout)
	return w.layout
}

// SetLayout set the split layout
func (w *Window) SetLayout(wl *WndLayout) error {
	if wl == nil {
		return ErrBadParams
	}
	w.MakeCurrent()
	if err := treeFixLoaded(wl, w.Self.(IWindow)); err != nil {
		return err
	}
	w.layout = wl
	return nil
}

// State to string
func (w *Window) State() ([]byte, error) {
	return w.Layout().State()
}

// SetState from string
func (w *Window) SetState(data []byte) error {
	wl := new(WndLayout)
	if err := wl.SetState(data); err != nil {
		return err
	}
	return w.SetLayout(wl)
}

// Render the scene
func (w *Window) Render() {
	glman.StackMatM.Push()
	glman.StackMatM.Load(geom.Mat4Ident())
	glman.StackMatV.Push()
	glman.StackMatV.Load(w.matView)
	glman.StackMatP.Push()
	glman.StackMatP.Load(w.matProj)

	defer func() {
		glman.StackMatM.Pop()
		glman.StackMatV.Pop()
		glman.StackMatP.Pop()
	}()

	//dbg.Logln("func (w *Window) Render()")
	w.MakeCurrent()
	//gl.ClearColor(0.8, 0.8, 0.9, 1.0)
	gl.ClearColor(0, 0, 0, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	if w.layout != nil {
		w.layout.Render(func(pn IPane) bool { return pn.Is3D() })
		w.layout.Render(func(pn IPane) bool { return !pn.Is3D() })
	}

	glman.DynDrawRect(Rect{10, 10, 300, 300}, Color{0, 0, 1, 0.5}, 3)
	glman.DynDrawText("ASDF", Rect{10, 60, 300, 300}, glman.LoadFont("WQY-ZenHei", 20), Color{0, 0, 1, 1}, 0)

	w.Present()

	// TODO: move to proper location
	runtime.GC()
	glman.Routine()
}
