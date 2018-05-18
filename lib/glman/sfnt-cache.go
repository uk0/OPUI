package glman

import (
	"log"
	"math"
	"strings"
	"tetra/internal/jurafont"
	"tetra/internal/refc"
	"tetra/lib/dbg"
	"tetra/lib/levenshtein"
	"tetra/lib/store"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var (
	sfcache  = make(map[string]*exSfnt)
	fntfiles map[string]bool
)

// multi font in one struct
type exSfnt struct {
	refc.Obj
	sfs     []*sfnt.Font
	metrics [][3]fixed.Int26_6
}

// NumGlyphs returns the number of glyphs in f.
func (f *exSfnt) NumGlyphs() (n int) {
	for _, p := range f.sfs {
		n += p.NumGlyphs()
	}
	return
}

// GlyphIndex returns the glyph index for the given rune.
//
// It returns (0, nil) if there is no glyph for r.
// https://www.microsoft.com/typography/OTSPEC/cmap.htm says that "Character
// codes that do not correspond to any glyph in the font should be mapped to
// glyph index 0. The glyph at this location must be a special glyph
// representing a missing character, commonly known as .notdef."
func (f *exSfnt) GlyphIndex(b *sfnt.Buffer, r rune) (uint32, error) {
	var nodef bool
	for i, p := range f.sfs {
		if x, err := p.GlyphIndex(b, r); err == nil {
			if x == 0 {
				nodef = true
			}
			return uint32(i)<<16 | uint32(x), nil
		}
	}
	if nodef {
		return 0, nil
	}
	return 0, sfnt.ErrNotFound
}

// LoadGlyph returns the vector segments for the x'th glyph. ppem is the number
// of pixels in 1 em.
//
// If b is non-nil, the segments become invalid to use once b is re-used.
//
// In the returned Segments' (x, y) coordinates, the Y axis increases down.
//
// It returns ErrNotFound if the glyph index is out of range. It returns
// ErrColoredGlyph if the glyph is not a monochrome vector glyph, such as a
// colored (bitmap or vector) emoji glyph.
func (f *exSfnt) LoadGlyph(b *sfnt.Buffer, x uint32, ppem fixed.Int26_6, opts *sfnt.LoadGlyphOptions) ([]sfnt.Segment, error) {
	i := int(x >> 16)
	y := sfnt.GlyphIndex(x & 0xFFFF)
	if i >= len(f.sfs) {
		return nil, sfnt.ErrNotFound
	}
	return f.sfs[i].LoadGlyph(b, y, ppem, opts)
}

// GlyphAdvance returns the advance width for the x'th glyph. ppem is the
// number of pixels in 1 em.
//
// It returns ErrNotFound if the glyph index is out of range.
func (f *exSfnt) GlyphAdvance(b *sfnt.Buffer, x uint32, ppem fixed.Int26_6, h font.Hinting) (fixed.Int26_6, error) {
	i := int(x >> 16)
	y := sfnt.GlyphIndex(x & 0xFFFF)
	if i >= len(f.sfs) {
		return 0, sfnt.ErrNotFound
	}
	return f.sfs[i].GlyphAdvance(b, y, ppem, h)
}

// // Metrics return metrics
// func (f *exSfnt) Metrics(b *sfnt.Buffer, x uint32, ppem fixed.Int26_6, h font.Hinting) (ascender, descender, lineGap fixed.Int26_6) {
// 	if len(f.metrics) == 0 {
// 		for _, sf := range f.sfs {
// 			a, d, g := sf.TetraMetrics(b, ppem, h)
// 			f.metrics = append(f.metrics, [3]fixed.Int26_6{a, d, g})
// 		}
// 	}
// 	i := int(x >> 16)
// 	if i >= len(f.metrics) {
// 		return 0, 0, 0
// 	}
// 	return f.metrics[i][0], f.metrics[i][1], f.metrics[i][2]
// }

func (f *exSfnt) finalize() {
	dbg.Logf("exSfnt.finalize()\n")
}

func finalizeSfnt(f *exSfnt) {
	var names []string
	for k, v := range sfcache {
		if v == f {
			names = append(names, k)
		}
	}
	for _, name := range names {
		delete(sfcache, name)
	}
	f.finalize()
}

func newExSfnt(sfs []*sfnt.Font) *exSfnt {
	f := new(exSfnt)
	f.sfs = sfs
	refc.SetFinalizer(&f.Obj, func() {
		finalizeSfnt(f)
	})
	return f
}

func parseSfntC(name string, b []byte) ([]*sfnt.Font, error) {
	c, err := sfnt.ParseCollection(b)
	if err != nil {
		return nil, err
	}
	var x []*sfnt.Font
	for i := 0; i < c.NumFonts(); i++ {
		if sf, err := c.Font(i); err == nil {
			x = append(x, sf)
			continue
		}
		log.Printf("warning: failed load %d font of collection %s\n", i, name)
	}
	return x, nil
}

// "arial bold" => "Airal(Bold)"
func matchFntFile(name string) (m string) {
	if fntfiles == nil {
		fntfiles = make(map[string]bool)
		ents, err := store.ReadDir("font")
		if err == nil {
			for _, info := range ents {
				x := info.Name()
				if info.IsDir() || !strings.HasSuffix(x, ".ttf") {
					continue
				}
				x = strings.TrimSpace(strings.TrimSuffix(x, ".ttf"))
				fntfiles[x] = true
			}
		}
		fntfiles["Default"] = true
	}
	if _, ok := fntfiles[name]; ok {
		return name
	}
	d := int(math.MaxInt32)
	for s := range fntfiles {
		d1 := levenshtein.DistanceCI(s, name)
		if d1 < d {
			m = s
			d = d1
		}
	}
	return
}

// load font file, if failed fallback to a embeded font
func loadSfnt(f Font) (sf *exSfnt, name string) {
	name = matchFntFile(f.Name())
	if sf, _ = sfcache[name]; sf != nil {
		return
	}
	b, err := store.ReadFile("font/" + name + ".ttf")
	if err == nil {
		var x []*sfnt.Font
		if x, err = parseSfntC(name, b); err == nil {
			sf = newExSfnt(x)
			sfcache[name] = sf
			return
		}
		log.Printf("warning: error parse ttf/otf: %v\n", err)
	} else {
		log.Printf("warning: %v\n", err)
	}
	if sf, _ = sfcache["Default"]; sf != nil {
		sfcache[name] = sf
		return
	}

	if x, err := sfnt.Parse(jurafont.TTF); err != nil {
		log.Panic("parse fallback font:", err)
	} else {
		sf = newExSfnt([]*sfnt.Font{x})
		sfcache[name] = sf
		sfcache["Default"] = sf
	}
	return
}
