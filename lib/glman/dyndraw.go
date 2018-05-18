package glman

import (
	"tetra/internal/gl"
)

var (
	dynArray12 *Res // 4x3 float
	dynArray20 *Res // 4x5 float
	dynArray30 *Res // 9x3 float
)

func bindDynArray12() {
	if dynArray12 == nil {
		dynArray12 = GenBuffer("*painter.dynArray12")
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, dynArray12.ID())
	DbgCheckError()
	gl.BufferData(gl.ARRAY_BUFFER, 12*4, nil, gl.STREAM_DRAW) // make orphan
	DbgCheckError()
}

func bindDynArray20() {
	if dynArray20 == nil {
		dynArray20 = GenBuffer("*painter.dynArray20")
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, dynArray20.ID())
	DbgCheckError()
	gl.BufferData(gl.ARRAY_BUFFER, 20*4, nil, gl.STREAM_DRAW) // make orphan
	DbgCheckError()
}

func bindDynArray30() {
	if dynArray30 == nil {
		dynArray30 = GenBuffer("*painter.dynArray30")
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, dynArray30.ID())
	DbgCheckError()
	gl.BufferData(gl.ARRAY_BUFFER, 30*4, nil, gl.STREAM_DRAW) // make orphan
	DbgCheckError()
}

// DynFillRect fill rectangle
func DynFillRect(rect Rect, color Color) {
	p := UseProgSimpleDraw()
	bindDynArray12()
	gl.Uniform4fv(p.UniColors, 1, &color[0])
	DbgCheckError()
	tmp := [12]float32{
		rect.X0(), rect.Y1(), 0,
		rect.X0(), rect.Y0(), 0,
		rect.X1(), rect.Y1(), 0,
		rect.X1(), rect.Y0(), 0}
	gl.BufferData(gl.ARRAY_BUFFER, 12*4, gl.Ptr(&tmp[0]), gl.STREAM_DRAW)
	DbgCheckError()
	gl.VertexAttribPointer(uint32(p.AttPos), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	DbgCheckError()
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(4))
	DbgCheckError()
}

// DynDrawRect draw rectangle
func DynDrawRect(rect Rect, color Color, lineWidth float32) {
	DynDrawRectEx(rect, color, lineWidth, lineWidth, lineWidth, lineWidth)
}

// DynDrawRectEx draw rectangle
func DynDrawRectEx(rect Rect, color Color, szLeft, szRight, szTop, szBottom float32) {
	p := UseProgSimpleDraw()
	bindDynArray30()
	gl.Uniform4fv(p.UniColors, 1, &color[0])
	DbgCheckError()

	//   x0    xa           xb     x1
	//   7------+-----------+------5 y0
	//   |      |           /  /   |
	//   +--\---6-----------4------+ ya
	//   |      |           |      |
	//   |      |           |      |
	//   0------8-----------2---\--+ yb
	//   |      /           |  \   |
	//   1------+-----------+------3 y1
	//
	x0, y0, x1, y1 := rect.X0(), rect.Y0(), rect.X1(), rect.Y1()
	xa, xb, ya, yb := x0+szLeft, x1-szRight, y0+szTop, y1-szBottom

	tmp := [30]float32{
		x0, yb, 0,
		x0, y1, 0,
		xb, yb, 0,
		x1, y1, 0,
		xb, ya, 0,
		x1, y0, 0,
		xa, ya, 0,
		x0, y0, 0,
		xa, yb, 0,
		x0, yb, 0}
	gl.BufferData(gl.ARRAY_BUFFER, 30*4, gl.Ptr(&tmp[0]), gl.STREAM_DRAW)
	DbgCheckError()
	gl.VertexAttribPointer(uint32(p.AttPos), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	DbgCheckError()
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(10))
	DbgCheckError()
}

// DynDrawImageRect fill rectangle
func DynDrawImageRect(rect Rect, color Color) {
	panic(nil)
	p := UseProgSimpleDraw()
	bindDynArray12()
	gl.Uniform4fv(p.UniColors, 1, &color[0])
	DbgCheckError()
	tmp := [12]float32{
		rect.X0(), rect.Y1(), 0,
		rect.X0(), rect.Y0(), 0,
		rect.X1(), rect.Y1(), 0,
		rect.X1(), rect.Y0(), 0}
	gl.BufferData(gl.ARRAY_BUFFER, 12*4, gl.Ptr(&tmp[0]), gl.STREAM_DRAW)
	DbgCheckError()
	gl.VertexAttribPointer(uint32(p.AttPos), 3, gl.FLOAT, false, 0, gl.PtrOffset(0))
	DbgCheckError()
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(4))
	DbgCheckError()
}

// OptionDrawText is options for draw text
type OptionDrawText int

// options for draw text
const (
	DtTop        OptionDrawText = 0x00000000
	DtLeft       OptionDrawText = 0x00000000
	DtCenter     OptionDrawText = 0x00000001
	DtRight      OptionDrawText = 0x00000002
	DtVCenter    OptionDrawText = 0x00000004
	DtBottom     OptionDrawText = 0x00000008
	DtSingleLine OptionDrawText = 0x00000010
)

// DynDrawText draw text
func DynDrawText(s string, rect Rect, font Font, color Color, options OptionDrawText) {
	p := UseProgTexFont(false)
	f := accessFont(font)
	//bindDynArray30()
	gl.Uniform4fv(p.UniColors, 1, &color[0])
	DbgCheckError()
	gl.Uniform1f(p.UniTexSize, float32(f.texsize))
	DbgCheckError()

	bindDynArray20()
	gl.ActiveTexture(gl.TEXTURE0)
	x0 := rect.X0()
	y0 := rect.Y0() + f.lineGap
	y1 := rect.Y0() + float32(f.height)
	for _, ch := range s {
		g := f.loadGlyph(ch)
		x1 := x0 + float32(g.w)

		// padding 1 pixel is inluded, we use normalized space because the lack of texelFetch func
		tx0 := float32(g.x) / float32(f.texsize)
		tx1 := float32(g.x+g.w) / float32(f.texsize)
		ty0 := (float32(f.ppem+2) * float32(g.row)) / float32(f.texsize)
		ty1 := (float32(f.ppem+2) * float32(g.row+1)) / float32(f.texsize)

		v := [4][5]float32{
			{x0, y1, 0, tx0, ty1}, // front face is CCW
			{x1, y1, 0, tx1, ty1}, //  2 3
			{x0, y0, 0, tx0, ty0}, //  |\|
			{x1, y0, 0, tx1, ty0}, //  0 1
		}
		gl.BufferData(gl.ARRAY_BUFFER, 20*4, nil, gl.STREAM_DRAW) // make orphan
		DbgCheckError()
		gl.BufferData(gl.ARRAY_BUFFER, 20*4, gl.Ptr(&v[0][0]), gl.STREAM_DRAW)
		DbgCheckError()
		gl.VertexAttribPointer(uint32(p.AttPos), 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
		DbgCheckError()
		gl.VertexAttribPointer(uint32(p.AttTC), 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
		DbgCheckError()
		gl.BindTexture(gl.TEXTURE_2D, f.textures[g.tex].ID())
		DbgCheckError()
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, int32(4))
		DbgCheckError()

		x0 = x1 - 2 //
	}
}
