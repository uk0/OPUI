package winl

// #cgo darwin LDFLAGS: -framework Cocoa
// #cgo windows LDFLAGS: -lgdi32 -lopengl32 -lglu32
// #cgo linux LDFLAGS: -lX11 -lGL -lGLU
// #include <stdlib.h>
// #include "winl-c.h"
import "C"

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"tetra/lib/dbg"
	"unsafe"

	"tetra/internal/gl"
)

//go:generate go run ../../cmd/classp/classp.go .

var (
	// map C window to Go window
	winMap = make(map[C.NativeWnd]*Window)
)

type hints uint32

const (
	// HintResizable is hint for window has resize box
	HintResizable hints = C.WINL_HINT_RESIZABLE

	// HintVideo is hint for drawing animated (FPS video) content
	HintVideo hints = C.WINL_HINT_ANIMATE

	// HintPainter is hint for 2D drawing using painter algorithm
	HintPainter hints = C.WINL_HINT_PAINTER

	// Hint3D is hint for 3D drawing
	Hint3D hints = C.WINL_HINT_3D
)

var (
	nilwin   C.NativeWnd // value for not a window
	started  bool
	creating *Window
	exitErr  error
	exited   bool
	glinited bool
)

// Window class wrap operating systems's window object.
type Window struct {
	Self IWindow

	native C.NativeWnd
	width  float32
	height float32
	hints  hints
}

// Init the object
func (w *Window) Init() {
}

// OnCreate event handler
func (w *Window) OnCreate() {
	// dbg.Logf("OnCreate()\n")
}

// OnDestroy event handler
func (w *Window) OnDestroy() {
	// dbg.Logf("OnDestroy()\n")
}

// OnResize event handler
func (w *Window) OnResize(width, height float32) {
	// dbg.Logf("OnResize(%f, %f)\n", width, height)
}

// OnMouseEnter event handler
func (w *Window) OnMouseEnter(x, y float32) {
	dbg.Logf("OnMouseEnter(%f, %f)\n", x, y)
}

// OnMouseLeave event handler
func (w *Window) OnMouseLeave(x, y float32) {
	// dbg.Logf("OnMouseLeave(%f, %f)\n", x, y)
}

// OnMouseMove event handler
func (w *Window) OnMouseMove(x, y float32) {
	// dbg.Logf("OnMouseMove(%f, %f)\n", x, y)
}

// OnMousePress event handler
func (w *Window) OnMousePress(btn int, x, y float32) {
	// dbg.Logf("OnMousePress(%d, %f, %f)\n", btn, x, y)
}

// OnMouseRelease event handler
func (w *Window) OnMouseRelease(btn int, x, y float32) {
	// dbg.Logf("OnMouseRelease(%d, %f, %f)\n", btn, x, y)
}

// OnMouseWheel event handler
func (w *Window) OnMouseWheel(vert bool, dz float32) {
	// if vert {
	// 	dbg.Logf("OnMouseWheel(vert, %f)\n", dz)
	// } else {
	// 	dbg.Logf("OnMouseWheel(horz, %f)\n", dz)
	// }
}

// OnExpose event handler
func (w *Window) OnExpose(x, y, width, height float32) {
	// dbg.Logf("OnExpose(%g, %g, %g, %g)\n", x, y, width, height)
}

// SetHints set hints for window style
func (w *Window) SetHints(hints hints) {
	w.hints |= hints
}

// Create the window, width and height can be zero.
func (w *Window) Create(width, height int) error {
	if !started {
		//panic("can't create window before func Run()")
	}
	if w.native != nilwin {
		panic("window already created.")
	}
	if w.Self == nil {
		w.Self = w
	}
	creating = w
	defer func() {
		creating = nil
	}()
	w.native = C.winl_create(C.int(w.hints), C.int(width), C.int(height))
	if w.native == nilwin {
		panic("failed to crate native window.")
	}
	winMap[w.native] = w
	C.winl_make_current(w.native) // important
	if !glinited {
		if err := gl.Init(); err != nil {
			log.Fatalln("fatal error: gl.Init(): ", err)
		}
		glinited = true
		version := gl.GoStr(gl.GetString(gl.VERSION))
		log.Println("OpenGL version:", version)
	}
	w.Self.OnCreate()

	return nil
}

func cleanUpDestroyed(w *Window) {
	if w.native == nilwin {
		return
	}
	C.winl_make_current(nilwin)
	delete(winMap, w.native)
	w.native = nilwin
}

// Destroy the window
func (w *Window) Destroy() {
	if w.native != nilwin {
		C.winl_destroy(w.native)
		cleanUpDestroyed(w)
	}
}

// IsVisible determine if window is visible
func (w *Window) IsVisible() bool {
	if w.native == nilwin {
		return false
	}
	if C.winl_is_visible(w.native) == 0 {
		return false
	}
	return true
}

// IsFullScreen determine if window is full screen
func (w *Window) IsFullScreen() bool {
	if w.native == nilwin {
		return false
	}
	if C.winl_is_full_screen(w.native) == 0 {
		return false
	}
	return true
}

// ToggleFullScreen switch between full screen mode and normal mode
func (w *Window) ToggleFullScreen() {
	if w.native == nilwin {
		return
	}
	C.winl_toggle_full_screen(w.native)
	// C.winl_track_mouse(w.native, 1-C.winl_is_full_screen(w.native))
}

// Show the window
func (w *Window) Show() {
	C.winl_show(w.native)
}

// MakeCurrent set current OpenGL to this window
func (w *Window) MakeCurrent() bool {
	dbg.Logln("(w *Window) MakeCurrent() bool")
	return C.winl_make_current(w.native) != 0
}

// Present copy OpenGL content from back buffer to front buffer, make it visible
func (w *Window) Present() {
	C.winl_swap_buffers(w.native)
}

// Size reports size of window's client area
func (w *Window) Size() (width, height float32) {
	return w.width, w.height
}

// func (w *Window) MousePos() (x, y float32) {
// 	var cx, cy C.float
// 	C.winl_get_mouse_pos(w.native, &cx, &cy)
// 	return float32(cx), float32(cy)
// }

// SetTitle set the window title
func (w *Window) SetTitle(title string) {
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))
	C.winl_set_title(w.native, ctitle)
}

// Expose triggle expose event
func (w *Window) Expose(x, y, width, height float32) {
	C.winl_expose(w.native, C.float(x), C.float(y), C.float(width), C.float(height))
}

func goWin(win C.NativeWnd) *Window {
	return winMap[win]
}

//export winl_on_resize
func winl_on_resize(win C.NativeWnd, width, height float32) {
	w := goWin(win)
	if w == nil {
		w = creating
	}
	if w == nil {
		return
	}
	w.width, w.height = width, height
	w.Self.OnResize(width, height)
}

//export winl_on_destroy
func winl_on_destroy(win C.NativeWnd) {
	w := goWin(win)
	if w == nil {
		return
	}
	w.Self.OnDestroy()
	cleanUpDestroyed(w)
}

//export winl_on_mouse_move
func winl_on_mouse_move(win C.NativeWnd, x, y float32) {
	w := goWin(win)
	if w == nil {
		return
	}
	w.Self.OnMouseMove(x, y)
}

//export winl_on_mouse_press
func winl_on_mouse_press(win C.NativeWnd, btn C.int, x, y float32) {
	w := goWin(win)
	if w == nil {
		return
	}
	w.Self.OnMousePress(int(btn), x, y)
}

//export winl_on_mouse_release
func winl_on_mouse_release(win C.NativeWnd, btn C.int, x, y float32) {
	w := goWin(win)
	if w == nil {
		return
	}

	w.Self.OnMouseRelease(int(btn), x, y)
}

//export winl_on_mouse_wheel
func winl_on_mouse_wheel(win C.NativeWnd, vertical C.int, dz float32) {
	w := goWin(win)
	if w == nil {
		return
	}
	var vert bool
	if vertical != 0 {
		vert = true
	}
	w.Self.OnMouseWheel(vert, dz)
}

//export winl_on_mouse_enter
func winl_on_mouse_enter(win C.NativeWnd, x, y float32) {
	w := goWin(win)
	if w == nil {
		return
	}
	w.Self.OnMouseEnter(x, y)
}

//export winl_on_mouse_leave
func winl_on_mouse_leave(win C.NativeWnd, x, y float32) {
	w := goWin(win)
	if w == nil {
		return
	}
	w.Self.OnMouseLeave(x, y)
}

//export winl_on_expose
func winl_on_expose(win C.NativeWnd, x, y, width, height float32) {
	w := goWin(win)
	if w == nil {
		return
	}
	w.Self.OnExpose(x, y, width, height)
}

// ScreenSize return size of main screen
func ScreenSize() (width, height int) {
	var w, h C.int
	C.winl_get_screen_size(&w, &h)
	return int(w), int(h)
}

//export winl_report
func winl_report(msg *C.char, topanic C.int) {
	s := strings.TrimSpace(C.GoString(msg))
	dbg.Logln("winl-c:", s)
	if topanic != 0 {
		panic(s)
	}
}

var gOnStart func() error

//export winl_on_start
func winl_on_start() {
	if gOnStart == nil {
		return
	}
	if err := gOnStart(); err != nil {
		Exit(err)
	}
}

var gOnExit func(error)

//export winl_on_exit
func winl_on_exit(code C.int) {
	if exited {
		return
	}
	exited = true

	if gOnExit != nil {
		var err error
		if exitErr != nil {
			err = exitErr
		} else if code != 0 {
			err = fmt.Errorf("Exit code=%d", code)
		}
		gOnExit(err)
	}
}

// OSVersion return the version of operating system
func OSVersion() (s string) {
	buf := C.winl_os_version()
	s = C.GoString(buf)
	C.free(unsafe.Pointer(buf))
	return
}

// Run the main event loop
func Run(onStart func() error, onExit func(err error)) {
	if started {
		panic("re-enter func Run()")
	}
	started = true
	gOnStart = onStart
	gOnExit = onExit
	code := C.winl_event_loop()
	winl_on_exit(code)
}

// Exit the event loop
func Exit(err error) {
	dbg.Log("exit event loop with err = ", err)
	exitErr = err
	var code C.int
	if err != nil {
		code = 1
	}
	C.winl_exit_loop(code)
}

// List of all windows
func List() (ret []*Window) {
	for _, v := range winMap {
		ret = append(ret, v)
	}
	return
}

func init() {
	// create then destroy a window, force opengl init
	w := new(Window)
	w.Create(100, 100)

	if runtime.GOOS == "darwin" {
		w.Destroy()
	}
}

func messageBoxLinux(w *Window, msg, title string) {
	if !started {
		panic("can't popup window before func Run()")
	}
	path, err := exec.LookPath("gxmessage")
	if err == nil {
		cmd := exec.Command(path, "-center", "-geometry", "600x200",
			"-buttons", "GTK_STOCK_OK:0",
			"-title", title, msg)
		cmd.Run()
		return
	}

	path, err = exec.LookPath("xmessage")
	if err == nil {
		cmd := exec.Command(path, "-center", "-geometry", "600x200",
			"-buttons", "OK:0",
			"-title", title, msg)
		cmd.Run()
		return
	}
	log.Println("warning: failed to locate gxmessage or xmessage utility, unable to show message box.")
	path, err = exec.LookPath("notify-send")
	if err == nil {
		cmd := exec.Command(path, "Please install gxmessage or xmessage utility.")
		cmd.Run()
	}
}

func confirmBoxLinux(w *Window, msg, title string) bool {
	if !started {
		panic("can't popup window before func Run()")
	}
	path, err := exec.LookPath("gxmessage")
	if err == nil {
		cmd := exec.Command(path, "-center", "-geometry", "600x200",
			"-buttons", "GTK_STOCK_CANCEL:1,GTK_STOCK_OK:0",
			"-title", title, msg)
		if err = cmd.Run(); err == nil {
			return true
		}
		return false
	}

	path, err = exec.LookPath("xmessage")
	if err == nil {
		cmd := exec.Command(path, "-center", "-geometry", "600x200",
			"-buttons", "Cancel:1,OK:0",
			"-title", title, msg)
		if err = cmd.Run(); err == nil {
			return true
		}
		return false
	}
	log.Println("wraning: failed to locate gxmessage or xmessage utility, unable to show message box.")
	path, err = exec.LookPath("notify-send")
	if err == nil {
		cmd := exec.Command(path, "Please install gxmessage or xmessage utility.")
		cmd.Run()
	}
	return false
}

// MessageBox show a simple message box
func MessageBox(w *Window, msg, title string) {
	//if !started {
	//	panic("can't popup window before func Run()")
	//}
	if runtime.GOOS == "linux" {
		messageBoxLinux(w, msg, title)
		return
	}
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))
	var nw C.NativeWnd
	if w != nil {
		nw = w.native
	}
	C.winl_message_box(nw, cmsg, ctitle)
}

// ConfirmBox show a simple box ask user to confirm.
func ConfirmBox(w *Window, msg, title string) bool {
	//if !started {
	//	panic("can't popup window before func Run()")
	//}
	if runtime.GOOS == "linux" {
		return confirmBoxLinux(w, msg, title)
	}
	cmsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cmsg))
	ctitle := C.CString(title)
	defer C.free(unsafe.Pointer(ctitle))
	var nw C.NativeWnd
	if w != nil {
		nw = w.native
	}
	if C.winl_confirm_box(nw, cmsg, ctitle) != 0 {
		return true
	}
	return false
}
