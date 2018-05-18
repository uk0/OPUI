package glman

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"tetra/lib/dbg"
	"tetra/lib/ssvg"
	"tetra/lib/store"

	"tetra/internal/gl"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

var (
	debug          = true
	maxFontTexSize = 128
	sfntBuffer     sfnt.Buffer // we don't need concurrent call, so use a golbal buffer
)

// Font descriptor, i.e."Arial(Bold) 20", "Monospace 24".
// if the program don't know such font, it will fallback to default font.
// it's discourage make Font from a string by hand, use LoadFont instead.
type Font string

// Name the font name, i.e. "Arial"
func (f Font) Name() string {
	s := string(f)
	if pos := strings.IndexByte(s, ' '); pos != -1 {
		s = s[:pos]
	}
	s = strings.TrimSpace(s)
	return s
}

// Size is the font size
func (f Font) Size() int {
	s := strings.TrimSpace(string(f))
	if pos := strings.LastIndexByte(s, ' '); pos != -1 {
		s = s[pos+1:]
	}
	if x, err := strconv.ParseInt(s, 10, 32); err == nil {
		return int(x)
	}
	return 20
}

// glyph position
type glyph struct {
	ch    rune
	x     uint16 // x position
	tex   uint8  // texture id
	row   uint8  // row number
	w     uint16 // glyph width
	count uint16 // reference count or release order
}

// sort by location
type glyphSlice []*glyph

func (p glyphSlice) Len() int           { return len(p) }
func (p glyphSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p glyphSlice) Less(i, j int) bool { return glyphLess(p[i], p[j]) }

func glyphLess(i, j *glyph) bool {
	if i.tex == j.tex {
		if i.row == j.row {
			return i.x < j.x
		}
		return i.row < j.row
	}
	return i.tex < j.tex
}

// sort by count
type idleSlice []*glyph

func (p idleSlice) Len() int           { return len(p) }
func (p idleSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p idleSlice) Less(i, j int) bool { return p[i].count < p[j].count }

// ---+-------------------
//    |        line gap
//    |  -+---------------
//    |   |
//    |   |
// height |
//    |  em      org
//    |   |--x--+
//    |   |     | y
// ---+---+---------------
type fontcfg struct {
	LineGap float32 `json:"linegap"`
	Origin  struct {
		X float32 `json:"x"`
		Y float32 `json:"y"`
	} `json:"origin"`
}

type texFont struct {
	sf *exSfnt // where glyphs load from

	name string // Arial

	ppem     int           // em is "the height of the type piece", ppem is pixel per em, thus row height
	ppemfx   fixed.Int26_6 // same as ppem
	orgX     float32
	orgY     float32
	height   float32
	lineGap  float32
	texsize  int             // width and height of textures, determine by the glyph count in font file
	textures []*Res          // textures
	free     glyphSlice      // free spaces, keep in sorted order
	alive    map[rune]*glyph // allocated glyphs
	idle     map[rune]*glyph // glyphs pool pending for free up
	iorder   uint16          // the order of free up, increase when put into idle
	fffd     *glyph          // sepcial glyph, U+FFFD REPLACEMENT CHARACTER

	svg *ssvg.Svg

	// abount the space allocation algorithms:
	//
	// our algorithm must handle 30,000+ characters of CJK characters at same time.
	// if we load all the characters as 20px square into memory, we need a 4k texture.
	// it seams acceptable for morden graphics card, but what if we need 80px glyphs?
	// 3 different font at same time? the dynamic space allocations may be a good choice.
	//
	// initial state of texture, split into rows of free space. each row has same height (ppem):
	// +---------------------------------+
	// | free space row 1                |
	// +---------------------------------+
	// | free space row 2                |
	// +---------------------------------+
	// | free space row 3                |
	// +---------------------------------+
	// +---------------------------------+
	//
	// after load 3 glyphs:
	// +---+---+---+---------------------+
	// | A | B | C |  free space row 1   |
	// +---+---+---+---------------------+
	// | free space row 2                |
	// +---------------------------------+
	// | free space row 3                |
	// +---------------------------------+
	// +---------------------------------+
	//
	// in the case release glyph 'B', put into idle, and free up later.
	// recover small piece of free space.
	// +---+---+---+---------------------+
	// | A |   | C |  free space row 1   |
	// +---+---+---+---------------------+
	// | free space row 2                |
	// +---------------------------------+
	// | free space row 3                |
	// +---------------------------------+
	// +---------------------------------+
	//
	// if we then release 'C', the free spaces become continious
	// +---+-----------------------------+
	// | A |     merged free space       |
	// +---+-----------------------------+
	// | free space row 2                |
	// +---------------------------------+
	// | free space row 3                |
	// +---------------------------------+
	// +---------------------------------+
	//
	// when the textures full, alloc another one
	// +---+---+---+---+---+---+---+---+-+   +---+---+---+---------------------+
	// | A | L | B | O | N | F | W | X | |   | Y | Z | a |                     |
	// +---+---+---+---+---+---+---+---+-+   +---+---+---+---------------------+
	// | I | K | R | C | S | M | H | U | |   |                                 |
	// +---+---+---+---+---+---+---+---+-+   +---------------------------------+
	// | J | D | Q | P | E | T | G | V | |   |                                 |
	// +---+---+---+---+---+---+---+---+-+   +---------------------------------+
	// +---------------------------------+   +---------------------------------+
	//
	// the free spaces is sorted in nature order, so the merge algorithm is efficient
	//
}

func (f *texFont) init(fn Font) {
	f.alive = make(map[rune]*glyph)
	f.idle = make(map[rune]*glyph)
	f.ppem = fn.Size()
	f.ppemfx = fixed.I(f.ppem)
	f.sf, f.name = loadSfnt(fn)
	f.sf.Retain()
	f.iorder = 65500
	dbg.Logf("%s num glyphs = %d\n", f.name, f.sf.NumGlyphs())

	var cfg fontcfg
	store.LoadState("font", f.name+".ttf", &cfg)
	f.orgX = cfg.Origin.X * float32(f.ppem)
	f.orgY = (1 - cfg.Origin.Y) * float32(f.ppem)
	f.height = float32(f.ppem+2) + float32(int(float32(f.ppem)*cfg.LineGap+0.5))
	f.lineGap = f.height - float32(f.ppem+2)

	// determine texture size, we assume glyph is square.
	// ascii fonts may only need a small size
	n := f.sf.NumGlyphs()
	ts := int(math.Sqrt(float64(n * f.ppem * f.ppem)))
	for f.texsize = 128; f.texsize < ts; f.texsize *= 2 {
		// nothing
	}
	if f.texsize >= maxFontTexSize {
		f.texsize = maxFontTexSize
	}

	if debug {
		f.svg = new(ssvg.Svg)
	}

	// setup U+FFFD, and ensure we can actually load glyph
	f.fffd = f.loadGlyphImg('\uFFFD')
	if f.fffd == nil {
		log.Panicf("failed to load U+FFFD for %s", f)
	}
	f.alive['\uFFFD'] = f.fffd
	f.fffd.count++
}

func (f *texFont) finalize() {
	dbg.Logf("(texFont %s).finalize()\n", f.name)
	if f.svg != nil {
		f.purge(true)
		f.svg.CurrentFrame().KeepVisible = true
		f.svg.WriteFile("_"+f.String()+".svg", 200)
	}
	// TODO: delete textures
	f.sf.Release()
}

// alloc texture and mark as free space
func (f *texFont) allocTexture() {
	texid := len(f.textures)
	texture := GenTexture(fmt.Sprintf("*font %s [%d]", f.name, texid))
	DbgCheckError()
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	DbgCheckError()
	gl.BindTexture(gl.TEXTURE_2D, texture.ID())
	DbgCheckError()
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	pixels := make([]byte, f.texsize*f.texsize)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.ALPHA8, int32(f.texsize), int32(f.texsize), 0,
		gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(pixels))
	DbgCheckError()

	f.textures = append(f.textures, texture)
	rows := f.texsize / (f.ppem + 2) // 1 pixel padding
	for i := 0; i < rows; i++ {
		f.free = append(f.free, &glyph{
			x: 0, tex: uint8(texid), row: uint8(i), w: uint16(f.texsize), count: 0,
		})
	}
	if f.svg != nil {
		f.svg.NextFrame()
		f.svg.CurrentFrame().KeepVisible = true
		f.svg.Add(&ssvg.Rect{X: 0, Y: float64(texid * f.texsize), W: float64(f.texsize), H: float64(f.texsize)})
	}
}

// alloc space for glyph
func (f *texFont) allocGlyph(w uint16) *glyph {
	//dbg.Logln("allocGlyph")
	start := 0
	for k := 0; k < 3; k++ {
		for i, g := range f.free[start:] {
			if g.w < w {
				continue
			}
			if g.w == w {
				f.free = append(f.free[:i], f.free[i+1:]...)
				return g
			}
			p := new(glyph)
			*p = *g
			g.w = g.w - w
			p.w = w
			g.x = g.x + w
			return p
		}
		if k == 0 {
			// purge and try again
			f.purge(false)
		} else if k == 1 {
			// alloc new texture and try again
			start = len(f.free)
			f.allocTexture()
		}
	}
	return nil
}

// load glyph image, return nil on failed, never fail for U+FFFD
func (f *texFont) loadGlyphImg(ch rune) *glyph {
	//dbg.Logln("loadGlyphImg")
	fallback := false // never fail for U+FFFD
	var adv fixed.Int26_6

	x, err := f.sf.GlyphIndex(&sfntBuffer, ch)
	if err != nil || x == 0 {
		if ch != '\uFFFD' {
			dbg.Logf("GlyphIndex: %#U not found", ch)
			return nil
		}
		fallback = true
		adv = fixed.I((f.ppem + 1) / 2)
	}
	var segments []sfnt.Segment
	if !fallback {
		segments, err = f.sf.LoadGlyph(&sfntBuffer, x, f.ppemfx, nil)
		if err != nil {
			dbg.Logf("LoadGlyph: %#U %v", ch, err)
			return nil
		}
		adv, err = f.sf.GlyphAdvance(&sfntBuffer, x, f.ppemfx, 0)
		if err != nil {
			adv = f.ppemfx
		}
	}
	if adv.Ceil() < 1 {
		adv = fixed.I(1) // our algorithm will failed when glyph.w == 0
	}
	width := adv.Ceil()
	height := f.ppem

	g := f.allocGlyph(uint16(width) + 2) // 1 pixel padding
	if g == nil {
		dbg.Logf("failed alloc for width=%d\n", width)
		return nil
	}
	g.count = 0
	g.ch = ch

	rect := image.Rect(0, 0, width, height)
	img := image.NewAlpha(rect)
	if fallback {
		a := color.Alpha{127}
		for x := 0; x < width; x++ {
			img.SetAlpha(x, 0, a)
			img.SetAlpha(x, height-1, a)
		}
		for y := 0; y < height; y++ {
			img.SetAlpha(0, y, a)
			img.SetAlpha(width-1, y, a)
		}
	} else {
		originX := f.orgX
		originY := f.orgY
		r := vector.NewRasterizer(width, height)
		r.DrawOp = draw.Src
		for _, seg := range segments {
			switch seg.Op {
			case sfnt.SegmentOpMoveTo:
				r.MoveTo(
					originX+float32(seg.Args[0].X)/64,
					originY+float32(seg.Args[0].Y)/64,
				)
			case sfnt.SegmentOpLineTo:
				r.LineTo(
					originX+float32(seg.Args[0].X)/64,
					originY+float32(seg.Args[0].Y)/64,
				)
			case sfnt.SegmentOpQuadTo:
				r.QuadTo(
					originX+float32(seg.Args[0].X)/64,
					originY+float32(seg.Args[0].Y)/64,
					originX+float32(seg.Args[1].X)/64,
					originY+float32(seg.Args[1].Y)/64,
				)
			case sfnt.SegmentOpCubeTo:
				r.CubeTo(
					originX+float32(seg.Args[0].X)/64,
					originY+float32(seg.Args[0].Y)/64,
					originX+float32(seg.Args[1].X)/64,
					originY+float32(seg.Args[1].Y)/64,
					originX+float32(seg.Args[2].X)/64,
					originY+float32(seg.Args[2].Y)/64,
				)
			}
		}
		r.Draw(img, rect, image.Opaque, image.Point{})
	}

	// padding 1 pixel around glyph
	pix := make([]byte, 0, int(g.w)*(f.ppem+2))
	v := img.Pix
	for i := 0; i < int(g.w); i++ {
		pix = append(pix, 0)
	}
	for j := 0; j < f.ppem; j++ {
		pix = append(pix, 0)
		pix = append(pix, v[:int(width)]...)
		pix = append(pix, 0)
		v = v[img.Stride:]
	}
	for i := 0; i < int(g.w); i++ {
		pix = append(pix, 0)
	}

	gl.BindTexture(gl.TEXTURE_2D, f.textures[g.tex].ID())
	DbgCheckError()
	gl.TexSubImage2D(gl.TEXTURE_2D, 0, int32(g.x), int32(g.row)*int32(f.ppem+2), int32(g.w), int32(f.ppem+2),
		gl.ALPHA, gl.UNSIGNED_BYTE, gl.Ptr(pix))

	return g
}

// load glyph, put into alive, fallback to U+FFFD on failed.
// don't increase reference count
func (f *texFont) loadGlyphNoRef(ch rune) *glyph {
	if f.svg != nil {
		defer f.dumpSvg()
	}
	g, ok := f.alive[ch]
	if ok {
		return g
	}
	g, ok = f.idle[ch]
	if ok {
		// dbg.Logf("reuse %#U\n", ch)
		delete(f.idle, ch)
		f.alive[ch] = g
		g.count = 0
		return g
	}
	if g = f.loadGlyphImg(ch); g != nil {
		f.alive[ch] = g
		return g
	}
	g = f.fffd
	return g
}

// load glyph, put into alive, fallback to U+FFFD on failed.
// increase reference count
func (f *texFont) loadGlyph(ch rune) (g *glyph) {
	g = f.loadGlyphNoRef(ch)
	g.count++
	return
}

// release reference of glyph
func (f *texFont) releaseGlyph(g *glyph) {
	if g.count == 0 {
		log.Panicf("over-release %#U", g.ch)
	}
	g.count--
	if g.count == 0 {
		f.putIdle(g)
	}
}

// put glyph into idle state, maybe unload or reuse later
func (f *texFont) putIdle(g *glyph) {
	// force purge when too many pending glyphs
	if len(f.idle) >= 18384 {
		f.purge(false)
	}
	// if iorder overflow, shift to useable
	if f.iorder == 0xFFFF {
		for _, g := range f.idle {
			g.count = g.count / 4
		}
		f.iorder = f.iorder / 4
	}

	delete(f.alive, g.ch)
	g.count = f.iorder
	f.iorder++
	f.idle[g.ch] = g
	//f.purge(true)
}

// dealloc about half of idle
func (f *texFont) purge(all bool) {
	if len(f.idle) == 0 {
		return
	}

	var s idleSlice
	for _, g := range f.idle {
		s = append(s, g)
	}
	sort.Sort(s)
	var n int
	if all {
		n = len(f.idle)
	} else {
		n = (len(f.idle) + 1) / 2
	}

	// put back remains
	f.idle = make(map[rune]*glyph)
	for _, g := range s[n:] {
		f.idle[g.ch] = g
	}

	// free up spaces
	for _, g := range s[:n] {
		f.dealloc(g)
	}

}

func canMerge(l, r *glyph) bool {
	if l.tex != r.tex || l.row != r.row {
		return false
	}
	if l.x+l.w > r.x {
		dbg.Logf("l = %+v\n", l)
		dbg.Logf("r = %+v\n", r)
		log.Panic("l.x+l.w > r.x")
	}
	return l.x+l.w >= r.x
}

// put back into free space
func (f *texFont) dealloc(g *glyph) {
	n := len(f.free)
	k := sort.Search(n, func(i int) bool {
		return glyphLess(g, f.free[i])
	})

	// ... [k-1] [k] ...
	//          ^
	//        insert
	var ml, mr bool
	var left, right *glyph
	if k > 0 {
		left = f.free[k-1]
		ml = canMerge(left, g)
	}
	if k < n {
		right = f.free[k]
		mr = canMerge(g, right)
	}

	if ml && mr {
		left.w = right.x + right.w - left.x
		tmp := f.free[k+1:]
		f.free = append(f.free[:k], tmp...)
	} else if ml {
		left.w += g.w
	} else if mr {
		right.x = g.x
		right.w += g.w
	} else {
		//g.count = 0
		//g.ch = 0
		var tmp []*glyph
		tmp = append(tmp, f.free[:k]...)
		tmp = append(tmp, g)
		tmp = append(tmp, f.free[k:]...)
		f.free = tmp
	}
}

func (f *texFont) dumpSvg() {
	if f.svg == nil {
		return
	}
	f.svg.NextFrame()
	f.svg.CurrentFrame().KeepVisible = false
	f.svg.FrameDuration = 100
	style := ssvg.Style{Transparency: 0.5, StrokeColor: "green", FillColor: "orange"}
	for _, g := range f.free {
		f.svg.Add(&ssvg.Rect{X: float64(g.x),
			Y: float64(g.tex)*float64(f.texsize) + float64(g.row)*float64(f.ppem),
			W: float64(g.w),
			H: float64(f.ppem), Style: style})
	}
	style1 := ssvg.Style{Transparency: 0.5, StrokeColor: "red", FillColor: "purple"}
	for _, g := range f.alive {
		f.svg.Add(&ssvg.Rect{X: float64(g.x),
			Y: float64(g.tex)*float64(f.texsize) + float64(g.row)*float64(f.ppem),
			W: float64(g.w),
			H: float64(f.ppem), Style: style1})
		f.svg.Add(&ssvg.Text{X: float64(g.x) + 3,
			Y:    float64(g.tex)*float64(f.texsize) + float64(g.row)*float64(f.ppem) + 3,
			Text: fmt.Sprintf("%c", g.ch),
		})
	}
	style2 := ssvg.Style{Transparency: 0.5, StrokeColor: "blue", FillColor: "cyan"}
	for _, g := range f.idle {
		f.svg.Add(&ssvg.Rect{X: float64(g.x),
			Y: float64(g.tex)*float64(f.texsize) + float64(g.row)*float64(f.ppem),
			W: float64(g.w),
			H: float64(f.ppem), Style: style2})
	}
	//f.svg.WriteFile("_"+f.String()+".svg", 300)
}

func (f *texFont) String() string {
	return fmt.Sprintf("[texFont %s]", f.name)
}

func (f *texFont) LoadString(s string) {
	defer func() {
		if e := recover(); e != nil {
			if f.svg != nil {
				f.svg.CurrentFrame().KeepVisible = true
				f.svg.WriteFile("_"+f.String()+".svg", 200)
			}
			panic(e)
		}
	}()
	var v []*glyph
	for _, ch := range s {
		g := f.loadGlyph(ch)
		v = append(v, g)
		if len(v) >= 10 {
			for _, x := range v {
				f.releaseGlyph(x)
			}
			v = nil
		}
	}
}

// LoadFont load a font, if failed it will match a near font
func LoadFont(name string, size int) Font {
	if size <= 0 {
		size = 20
	}
	if size < 6 {
		size = 6
	}
	name = matchFntFile(name)
	return Font(fmt.Sprintf("%s %d", name, size))
}
