// MIT License
//
// Copyright (c) 2016 C.T.Chen
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package ssvg is for generate simple SVG, support muti-frame animation.
package ssvg

import (
	"fmt"
	"io"
	"math"
	"os"
)

var colorTable = []string{
	"red",
	"yellow",
	"green",
	"blue",
	"cyan",
	"purple",
	"brown",
	"chocolate",
	"crimson",
	"darkgoldenrod",
	"darkkhaki",
	"darkslateblue",
	"darkslategrey",
	"darkred",
	"darkblue",
	"darkcyan",
	"goldenrod",
	"fuchsia",
	"deepskyblue",
	"darkorange",
	"orangered",
	"olive",
	"mediumseagreen",
	"midnightblue",
	"sandybrown",
	"lightcoral",
	"lightseagreen",
}

// DefaultColors is easy way to choose color for serial of items.
// index can be any nonnegative integer.
func DefaultColors(index int) string {
	return colorTable[index%len(colorTable)]
}

// Style keep styles for element.
type Style struct {
	// Fill color
	FillColor string

	// Stroke color
	StrokeColor string

	// Stroke width in millimetres, 0 is treat as 1
	StrokeWidth float64

	// Transparent level between 0~1, opposite to opacity.
	Transparency float64
}

// Write style attributes to w, for svg.
func (style *Style) Write(w io.Writer, svg *Svg) {
	fc := style.FillColor
	sc := style.StrokeColor
	sw := style.StrokeWidth
	op := 1 - style.Transparency
	if fc == "" {
		fc = "none"
	}
	if sc == "" {
		sc = "#000000"
	}
	if sw <= 0 {
		sw = 1
	}
	if op > 0 {
		fmt.Fprintf(w, ` style="fill:`+fc+
			`;stroke:`+sc+
			`;stroke-width:%g" opacity="%g"`, sw*svg.unit, op)
	} else {
		fmt.Fprintf(w, ` style="fill:`+fc+
			`;stroke:`+sc+
			`;stroke-width:%g" `, sw*svg.unit)
	}
}

// Element is abstract element.
type Element interface {
	Write(w io.Writer, svg *Svg)
	Range(xmin, xmax, ymin, ymax *float64)
}

// 折线等元素中的点, 如需"点状图元"请用Icon
type Point struct {
	X, Y float64
}

//func (this * Point) Write(w * io.Writer, unit, pr float64) {
//	s := fmt.Sprintf("<circle cx=\"%g\" cy=\"%g\" r=\"%g\" ", this.X, this.Y, pr)
//	io.WriteString(w, s)
//	Style.Write(w, svg)
//	io.WriteString(w, " /> \n")
//	break;
//}

func include(xmin, xmax, ymin, ymax *float64, x, y float64) {
	*xmin = math.Min(*xmin, x)
	*xmax = math.Max(*xmax, x)
	*ymin = math.Min(*ymin, y)
	*ymax = math.Max(*ymax, y)
}

// 简易图标, 大小是固定的
type Icon struct {
	X, Y  float64
	Shape string  // "box", "circle"
	Zoom  float64 // 0 ~ 1.0
	Style
}

func (this *Icon) _infSize() {
}

func (this *Icon) Range(xmin, xmax, ymin, ymax *float64) {
	// 无需考虑宽度, 因为输出的图形本来就留有边界
	include(xmin, xmax, ymin, ymax, this.X, this.Y)
}

func (this *Icon) Write(w io.Writer, svg *Svg) {
	is := svg.iconSize
	if this.Zoom != 0 {
		is *= this.Zoom
	}
	r := is * 0.5
	switch this.Shape {
	case "circle":
		p := &Circle{Cx: this.X, Cy: this.Y, R: r, Style: this.Style}
		p.Write(w, svg)
	default:
		fallthrough
	case "box":
		p := &Rect{X: this.X - r, Y: this.Y - r, W: is, H: is, Style: this.Style}
		p.Write(w, svg)
	case "x":
		fallthrough
	case "cross":
		p := &Line{X1: this.X - r, Y1: this.Y - r, X2: this.X + r, Y2: this.Y + r, Style: this.Style}
		p.Write(w, svg)
		p = &Line{X1: this.X - r, Y1: this.Y + r, X2: this.X + r, Y2: this.Y - r, Style: this.Style}
		p.Write(w, svg)
	case "+":
		fallthrough
	case "plus":
		p := &Line{X1: this.X, Y1: this.Y - r, X2: this.X, Y2: this.Y + r, Style: this.Style}
		p.Write(w, svg)
		p = &Line{X1: this.X - r, Y1: this.Y, X2: this.X + r, Y2: this.Y, Style: this.Style}
		p.Write(w, svg)

	}
}

type Line struct {
	X1, Y1, X2, Y2 float64
	AuxLeft        bool // 左侧辅助线
	AuxRight       bool // 右侧辅助线
	Arrow          bool // 箭头
	Style
}

func (this *Line) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, this.X1, this.Y1)
	include(xmin, xmax, ymin, ymax, this.X2, this.Y2)
}

func (this *Line) Write(w io.Writer, svg *Svg) {
	if (this.Arrow || this.AuxLeft || this.AuxRight) &&
		(this.X1 != this.X2 || this.Y1 != this.Y2) {
		offset := svg.pixSize * 1.5

		// 垂直方向(逆时针90度)
		dx, dy := this.Y1-this.Y2, this.X2-this.X1
		a := math.Sqrt(dx*dx + dy*dy)
		dx, dy = offset*dx/a, offset*dy/a
		m := *this
		if m.StrokeWidth <= 0 {
			m.StrokeWidth = 0.3
		} else {
			m.StrokeWidth *= 0.3
		}
		m.AuxLeft = false
		m.AuxRight = false
		m.Arrow = false

		if this.AuxLeft {
			m.X1, m.Y1 = this.X1+dx, this.Y1+dy
			m.X2, m.Y2 = this.X2+dx, this.Y2+dy
			m.Write(w, svg)
		}
		if this.AuxRight {
			m.X1, m.Y1 = this.X1-dx, this.Y1-dy
			m.X2, m.Y2 = this.X2-dx, this.Y2-dy
			m.Write(w, svg)
		}
		if this.Arrow {
			m.StrokeWidth = this.StrokeWidth
			dx1, dy1 := this.X2-this.X1, this.Y2-this.Y1
			dx1, dy1 = svg.iconSize*dx1/a, svg.iconSize*dy1/a

			m.X2, m.Y2 = this.X2, this.Y2
			m.X1, m.Y1 = this.X2-dx1+dx*1, this.Y2-dy1+dy*1
			m.Write(w, svg)
			m.X1, m.Y1 = this.X2-dx1-dx*1, this.Y2-dy1-dy*1
			m.Write(w, svg)
			//m.X1, m.Y1 = this.X1+dx, this.Y1+dy
			//m.X2, m.Y2 = this.X2+dx, this.Y2+dy
			//m.Write(w, svg)

		}
	}

	fmt.Fprintf(w, "<line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\" ", this.X1, this.Y1, this.X2, this.Y2)
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")
}

type HLine struct {
	Y float64
	Style
}

func (this *HLine) _infSize() {
}

func (this *HLine) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, *xmin, this.Y)
}

func (this *HLine) Write(w io.Writer, svg *Svg) {
	fmt.Fprintf(w, "<line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\" ", svg.xmin, this.Y, svg.xmax, this.Y)
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")
}

type VLine struct {
	X float64
	Style
}

func (this *VLine) _infSize() {
}

func (this *VLine) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, this.X, *xmin)
}

func (this *VLine) Write(w io.Writer, svg *Svg) {
	fmt.Fprintf(w, "<line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\" ", this.X, svg.ymin, this.X, svg.ymax)
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")
}

type Polygon struct {
	Points []Point
	Style
}

func (this *Polygon) Range(xmin, xmax, ymin, ymax *float64) {
	for _, pt := range this.Points {
		include(xmin, xmax, ymin, ymax, pt.X, pt.Y)
	}
}

func (this *Polygon) Write(w io.Writer, svg *Svg) {
	io.WriteString(w, "<polygon points=\"")
	for _, pt := range this.Points {
		fmt.Fprintf(w, "%g,%g ", pt.X, pt.Y)
	}
	io.WriteString(w, "\" ")
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")
}

type Polyline struct {
	Points []Point
	Style
}

func (this *Polyline) Range(xmin, xmax, ymin, ymax *float64) {
	for _, pt := range this.Points {
		include(xmin, xmax, ymin, ymax, pt.X, pt.Y)
	}
}

func (this *Polyline) Write(w io.Writer, svg *Svg) {
	io.WriteString(w, "<polyline points=\"")
	for _, pt := range this.Points {
		fmt.Fprintf(w, "%g,%g ", pt.X, pt.Y)
	}
	io.WriteString(w, "\" ")
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")
}

type Circle struct {
	Cx, Cy, R float64
	Style
}

func (this *Circle) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, this.Cx-this.R, this.Cy-this.R)
	include(xmin, xmax, ymin, ymax, this.Cx+this.R, this.Cy+this.R)
}

func (this *Circle) Write(w io.Writer, svg *Svg) {
	fmt.Fprintf(w, "<circle cx=\"%g\" cy=\"%g\" r=\"%g\" ", this.Cx, this.Cy, this.R)
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")

}

type Rect struct {
	X, Y, W, H float64
	Style
}

func (this *Rect) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, this.X, this.Y)
	include(xmin, xmax, ymin, ymax, this.X+this.W, this.Y+this.H)
}

func (this *Rect) Write(w io.Writer, svg *Svg) {
	fmt.Fprintf(w, "<rect x=\"%g\" y=\"%g\" width=\"%g\" height=\"%g\" ", this.X, this.Y, this.W, this.H)
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")

}

type Ellipse struct {
	Cx, Cy, Rx, Ry float64
	Style
}

func (this *Ellipse) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, this.Cx-this.Rx, this.Cy-this.Ry)
	include(xmin, xmax, ymin, ymax, this.Cx+this.Rx, this.Cy+this.Ry)
}

func (this *Ellipse) Write(w io.Writer, svg *Svg) {
	fmt.Fprintf(w, "<ellipse cx=\"%g\" cy=\"%g\" rx=\"%g\" ry=\"%g\" ", this.Cx, this.Cy, this.Rx, this.Ry)
	this.Style.Write(w, svg)
	io.WriteString(w, " /> \n")
}

type Text struct {
	X, Y float64
	Text string
	Zoom float64
	Style
}

func (this *Text) Range(xmin, xmax, ymin, ymax *float64) {
	include(xmin, xmax, ymin, ymax, this.X, this.Y)
}

func (this *Text) Write(w io.Writer, svg *Svg) {
	if svg.YDown {
		fmt.Fprintf(w, "<text x=\"%g\" y=\"%g\" ", this.X, this.Y)
	} else {
		fmt.Fprintf(w, "<text transform=\" translate(%g,%g) scale(1, -1)\" ",
			this.X, this.Y)
	}

	fc := this.FillColor
	sc := this.StrokeColor
	sw := this.StrokeWidth
	fs := this.Zoom
	if fc == "" {
		fc = "#000000"
	}
	if sc == "" {
		sc = "none"
	}
	if fs == 0 {
		fs = 1
	}

	fmt.Fprintf(w, ` style="fill:`+fc+
		`;stroke:`+sc+
		`;stroke-width:%g`, sw*svg.unit)

	fmt.Fprintf(w, ";font-size:%gem\" >", fs*svg.unit)
	fmt.Fprintf(w, this.Text)
	fmt.Fprintf(w, "</text>\n")
}

type Frame struct {
	Elements    []Element
	Duration    int // 毫秒
	KeepVisible bool
}

// 画布范围(逻辑坐标)的确定是自动的, 但要先统计有限的图元, 再统计无限的图元
// 这个接口用来标记无限图元, 以保证它们在有限图元之后计算
type _infSize interface {
	_infSize()
}

func isAutoSize(e Element) bool {
	_, b := e.(_infSize)
	return b
}

func (this *Frame) Range1(xmin, xmax, ymin, ymax *float64) {
	for _, e := range this.Elements {
		if isAutoSize(e) {
			continue
		}
		e.Range(xmin, xmax, ymin, ymax)
	}
}

func (this *Frame) Range2(xmin, xmax, ymin, ymax *float64) {
	for _, e := range this.Elements {
		if isAutoSize(e) {
			e.Range(xmin, xmax, ymin, ymax)
		}
	}
}

func (this *Frame) Add(e Element) {
	this.Elements = append(this.Elements, e)
}

type Svg struct {
	CanvasSize int // pixels

	Frames     []*Frame
	iconSize   float64
	pixSize    float64
	xmin, xmax float64
	ymin, ymax float64
	unit       float64

	YDown bool

	FrameDuration int // default frame duration
}

//func (this *Svg) KeepLastFrameVisible() {
//	this.CurrentFrame().KeepVisible = true
//}
func (this *Svg) CurrentFrame() *Frame {
	if len(this.Frames) == 0 {
		return this.NextFrame()
	}
	return this.Frames[len(this.Frames)-1]
}

func (this *Svg) NextFrame() *Frame {
	f := new(Frame)
	//f.Duration = 333
	//if len(this.Frames) == 0 {
	//	f.KeepVisible = true
	//}
	this.Frames = append(this.Frames, f)
	return f
}

func (this *Svg) Add(e Element) {
	this.CurrentFrame().Add(e)
}

//	 , file);
func (this *Svg) WriteFile(filename string, canvasPixelSize float64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	this.Write(f, canvasPixelSize)
	return nil
}

func (this *Svg) Write(w io.Writer, canvasPixelSize float64) {
	fmt.Fprintf(w, "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\" ?> \n")
	fmt.Fprintf(w, "<!DOCTYPE svg PUBLIC \"-//W3C//DTD SVG 1.1//EN\" \n")
	fmt.Fprintf(w, "     \"http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd\"> \n")

	xmin := math.MaxFloat32
	ymin := math.MaxFloat32
	xmax := -math.MaxFloat32
	ymax := -math.MaxFloat32
	if len(this.Frames) == 0 {
		xmin, ymin, xmax, ymax = 0, 0, 0, 0
	} else {
		for _, f := range this.Frames {
			f.Range1(&xmin, &xmax, &ymin, &ymax)
		}
		for _, f := range this.Frames {
			f.Range2(&xmin, &xmax, &ymin, &ymax)
		}
	}

	dx := xmax - xmin
	dy := ymax - ymin
	if dx == 0 && dy == 0 {
		dx = 0.001
		dy = 0.001
	} else if dx == 0 && dy != 0 {
		dx = dy
	} else if dy == 0 && dx != 0 {
		dy = dx
	}

	var pw, ph float64
	if canvasPixelSize <= 0 {
		pw = dx
		ph = dy
	} else {
		if dx > dy {
			pw = canvasPixelSize
			ph = pw * dy / dx
		} else {
			ph = canvasPixelSize
			pw = ph * dx / dy
		}

	}
	this.xmin, this.xmax = xmin, xmax
	this.ymin, this.ymax = ymin, ymax
	//	rc := &Rect{X: xmin, Y: ymin, W: dx, H: dy,
	//		Style: Style{StrokeColor: "lightgray", StrokeWidth: 0.5}}

	const border = 5.0

	lbx := border * dx / pw
	lby := border * dx / pw

	pw += border * 2
	ph += border * 2
	xmin -= lbx
	xmax += lbx
	ymin -= lby
	ymax += lby
	dx = xmax - xmin
	dy = ymax - ymin

	diagonal := math.Sqrt(dx*dx + dy*dy)
	//	pr := diagonal / 200.0
	this.unit = 0.3 * diagonal / math.Sqrt(pw*pw+ph*ph) // about one pixel

	fmt.Fprintf(w, "<svg width=\"%dmm\" height=\"%dmm\" viewBox=\"0 0 %g %g\" \n    xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\"> \n",
		int(math.Ceil(pw)), int(math.Ceil(ph)), dx, dy)
	defer io.WriteString(w, "</svg>\n")

	if !this.YDown {
		fmt.Fprintf(w, "<g transform=\"scale(1, -1) translate(%g, %g)\" > \n", -xmin, -ymin-dy)
	} else {
		fmt.Fprintf(w, "<g transform=\"translate(%g, %g) \" > \n", -xmin, -ymin)
	}
	defer io.WriteString(w, "</g>\n")

	// 背景框
	this.iconSize = 0.3 * (lbx + lby)
	this.pixSize = 0.5 * (dx/pw + dy/ph)
	//rc.Write(w, this)

	if len(this.Frames) == 1 {
		for _, e := range this.Frames[0].Elements {
			e.Write(w, this)
		}
		return
	}
	dfd := this.FrameDuration
	if dfd == 0 {
		dfd = 1000
	}
	var begin int
	for _, f := range this.Frames {
		fd := f.Duration
		if fd == 0 {
			fd = dfd
		}
		io.WriteString(w, "<g visibility=\"hidden\">\n")
		fmt.Fprintf(w, "<set attributeName=\"visibility\" attributeType=\"CSS\" to=\"visible\" begin=\"%dms\" dur=\"%dms\" fill=\"freeze\" />\n",
			begin, fd)
		begin += fd
		if !f.KeepVisible {
			fmt.Fprintf(w, "<set attributeName=\"visibility\" attributeType=\"CSS\" to=\"hidden\" begin=\"%dms\" dur=\"1s\" fill=\"freeze\" />\n",
				begin)
		}
		for _, e := range f.Elements {
			e.Write(w, this)
		}
		io.WriteString(w, "</g>\n")
	}

}
