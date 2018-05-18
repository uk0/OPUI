package glman

import (
	"runtime"

	"tetra/internal/gl"
)

// MText is text model
type MText interface {
	Render()
	Colors() []Color
	SetColors(c ...Color)
	DrawEdge() bool
	SetDrawEdge(b bool)
	Text() string
}

type mText struct {
	s    string
	f    *texFont
	gs   []*glyph
	vbo  *Res
	segs [][3]uint32 // [0]=texture, [1]=offset, [3]=count
	clr  [2]Color
	edge bool
}

func finalizeMText(m *mText) {
	for _, g := range m.gs {
		m.f.releaseGlyph(g)
	}
}

func (m *mText) Text() string {
	return m.s
}

func (m *mText) String() string {
	return m.s
}

func (m *mText) Colors() []Color {
	return m.clr[:]
}

func (m *mText) SetColors(clrs ...Color) {
	for i, c := range clrs {
		m.clr[i].Copy(c)
	}
}

func (m *mText) DrawEdge() bool {
	return m.edge
}

func (m *mText) SetDrawEdge(b bool) {
	m.edge = b
}

func (m *mText) Render() {
	if len(m.segs) == 0 {
		return
	}
	p := UseProgTexFont(m.edge)

	gl.Uniform4fv(p.UniColors, 2, &m.clr[0][0])
	DbgCheckError()

	gl.Uniform1f(p.UniTexSize, float32(m.f.texsize))
	DbgCheckError()

	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo.ID())
	DbgCheckError()
	//gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
	//gl.PolygonMode( gl.FRONT_AND_BACK, gl.FILL );
	gl.EnableVertexAttribArray(uint32(p.AttPos))
	DbgCheckError()
	gl.VertexAttribPointer(uint32(p.AttPos), 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	DbgCheckError()
	gl.EnableVertexAttribArray(uint32(p.AttTC))
	DbgCheckError()
	gl.VertexAttribPointer(uint32(p.AttTC), 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	DbgCheckError()
	gl.ActiveTexture(gl.TEXTURE0)
	for _, seg := range m.segs {
		gl.BindTexture(gl.TEXTURE_2D, seg[0])
		DbgCheckError()
		gl.DrawArrays(gl.TRIANGLE_STRIP, int32(seg[1]), int32(seg[2]))
		DbgCheckError()
	}
}

func (f *texFont) mkMText(s string, width, height float32, options uint32) (m *mText) {
	m = new(mText)
	runtime.SetFinalizer(m, finalizeMText)
	m.f = f
	m.s = s
	m.clr[0] = [4]float32{1, 0, 0, 1} // default to black opaque
	m.clr[1] = [4]float32{0, 0, 0, 1}
	m.edge = true
	m.vbo = GenBuffer("*mText.vbo")
	DbgCheckError()
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vbo.ID())
	DbgCheckError()

	var vas = make([][][5]float32, len(f.textures)) // [texture][vertex][x,y,z,tx,ty]
	x0 := float32(0)
	y0 := f.lineGap
	y1 := float32(f.height)
	for _, ch := range s {
		g := f.loadGlyph(ch)
		m.gs = append(m.gs, g)
		x1 := x0 + float32(g.w)

		// padding 1 pixel is inluded, we use normalized space because the lack of texelFetch func
		tx0 := float32(g.x) / float32(f.texsize)
		tx1 := float32(g.x+g.w) / float32(f.texsize)
		ty0 := (float32(f.ppem+2) * float32(g.row)) / float32(f.texsize)
		ty1 := (float32(f.ppem+2) * float32(g.row+1)) / float32(f.texsize)

		v := [4][5]float32{
			{x0, y1, 0, tx0, ty1}, // front face is CW
			{x0, y0, 0, tx0, ty0}, //  1 3
			{x1, y1, 0, tx1, ty1}, //  |\|
			{x1, y0, 0, tx1, ty0}, //  0 2
		}
		for i := len(vas); i < len(f.textures); i++ {
			vas = append(vas, nil)
		}
		if len(vas[g.tex]) != 0 {
			vas[g.tex] = append(vas[g.tex], v[0]) // prepend degenerated triangle
		}
		vas[g.tex] = append(vas[g.tex], v[0], v[1], v[2], v[3], v[3]) // also append degenerated triangle

		x0 = x1 - 2 //
	}

	// merge into single array
	var mva [][5]float32
	for i, va := range vas {
		if len(va) == 0 {
			continue
		}
		off := uint32(len(mva))
		count := uint32(len(va))
		m.segs = append(m.segs, [3]uint32{f.textures[i].ID(), off, count})
		mva = append(mva, va...)
	}

	// store into VBO
	gl.BufferData(gl.ARRAY_BUFFER, len(mva)*5*4, gl.Ptr(mva), gl.STATIC_DRAW)
	DbgCheckError()

	return
}

// MkMText create text model
func (f Font) MkMText(s string, width, height float32, options uint32) MText {
	return accessFont(f).mkMText(s, width, height, options)
}
