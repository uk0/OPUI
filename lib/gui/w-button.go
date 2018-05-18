package gui

import (
	"tetra/lib/dbg"
	"tetra/lib/glman"
)

// Button is button widget
type Button struct {
	Widget
	fnt glman.Font
	txt glman.MText
}

// Font returns current font
func (btn *Button) Font() glman.Font {
	return btn.fnt
}

// SetFont set the font
func (btn *Button) SetFont(f glman.Font) {
	btn.fnt = f
	s := btn.Text()
	btn.txt = nil
	btn.SetText(s)
}

// Text label on the button
func (btn *Button) Text() string {
	if btn.txt == nil {
		return ""
	}
	return btn.txt.Text()
}

// SetText set the text label on the button
func (btn *Button) SetText(s string) {
	if s == btn.Text() {
		return
	}
	btn.txt = btn.fnt.MkMText(s, btn.Bounds().Width(), btn.Bounds().Height(), 0)
}

// Render the element
func (btn *Button) Render() {
	dbg.Logf("func (btn *Button) Render()\n")
	if btn.txt != nil {
		btn.txt.Render()
	}
}
