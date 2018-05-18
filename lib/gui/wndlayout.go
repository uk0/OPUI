package gui

import (
	"encoding/json"
	"tetra/lib/dbg"
	"tetra/lib/factory"
)

// WndLayout is tree structure split window into multipile panes
type WndLayout struct {
	rc Rect // bound rect calc base on split and parent.rc

	Vert  bool    `json:"vert,omitempty"`  // split direction
	Sp    float32 `json:"split,omitempty"` // split position
	Pane  IPane   `json:"-"`               // associated pane
	Class string  `json:"class,omitempty"` // pane' class
	Param string  `json:"param,omitempty"` // params for create pane

	L *WndLayout `json:"left,omitempty"`  // left or top child
	R *WndLayout `json:"right,omitempty"` // right or bottom child
}

// IsLeaf report whether wl is leaf node
func (wl *WndLayout) IsLeaf() bool {
	return wl.L == nil && wl.R == nil
}

func treeFixLoaded(wl *WndLayout, w IWindow) error {
	if wl.IsLeaf() {
		if wl.Pane != nil {
			return nil
		}
		var ok bool
		if ctor := factory.Get(wl.Class); ctor == nil {
			dbg.Logf("factory method for \"%s\" not found, fallback to gui.Pane", wl.Class)
			wl.Pane = NewPane()
		} else if wl.Pane, ok = ctor().(IPane); !ok {
			dbg.Logf("returns of factory method of \"%s\" is not a Pane, fallback to gui.Pane", wl.Class)
			wl.Pane = NewPane()
		}
		wl.Pane.SetState([]byte(wl.Param))
		wl.Pane.SetWindow(w)
	}

	if wl.L != nil {
		if err := treeFixLoaded(wl.L, w); err != nil {
			return err
		}
	}
	if wl.R != nil {
		if err := treeFixLoaded(wl.R, w); err != nil {
			return err
		}
	}
	return nil
}

func treePrepairSave(wl *WndLayout) {
	if wl == nil {
		return
	}
	treePrepairSave(wl.L)
	treePrepairSave(wl.R)
	if wl.Pane != nil {
		wl.Class = wl.Pane.Class()
		b, _ := wl.Pane.State()
		wl.Param = string(b)
	}
}

// State to string
func (wl *WndLayout) State() ([]byte, error) {
	return json.MarshalIndent(wl, "", "  ")
}

// SetState from string
func (wl *WndLayout) SetState(data []byte) error {
	return json.Unmarshal(data, wl)
}

// CalcLayout recursive calc wl.rc, and set to wl.Pane if any
func (wl *WndLayout) CalcLayout(ss float32) {
	if wl.L == nil && wl.R == nil {
		if wl.Pane != nil {
			wl.Pane.SetBounds(wl.rc)
		}
		return
	}
	if wl.L != nil && wl.R == nil {
		wl.L.rc = wl.rc
		wl.L.CalcLayout(ss)
		return
	}
	if wl.L == nil && wl.R != nil {
		wl.R.rc = wl.rc
		wl.R.CalcLayout(ss)
		return
	}
	wl.L.rc = wl.rc
	wl.R.rc = wl.rc
	if wl.Vert {
		pos := wl.rc[1] + wl.rc.Height()*wl.Sp
		wl.L.rc[1] = wl.rc[1]
		wl.L.rc[3] = pos - ss*0.5
		wl.R.rc[1] = wl.L.rc[3] + ss
		wl.R.rc[3] = wl.rc[3]
	} else {
		pos := wl.rc[0] + wl.rc.Width()*wl.Sp
		wl.L.rc[0] = wl.rc[0]
		wl.L.rc[2] = pos - ss*0.5
		wl.R.rc[0] = wl.L.rc[2] + ss
		wl.R.rc[2] = wl.rc[2]
	}
	wl.L.CalcLayout(ss)
	wl.R.CalcLayout(ss)
}

// Render the layout tree
func (wl *WndLayout) Render(filter func(IPane) bool) {
	if wl.L != nil {
		wl.L.Render(filter)
	}
	if wl.R != nil {
		wl.R.Render(filter)
	}
	if wl.Pane != nil && filter(wl.Pane) {
		dbg.Logf("wl.Pane.Render(): %s\n", wl.Pane.Class())
		wl.Pane.Render()
	}
}
