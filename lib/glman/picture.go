package glman

import (
	"tetra/internal/gl"
	"unsafe"
)

const sizePackTex = 256

var (
	packTextures []*packTex
	//fallbackPic  = Image{width: 16, height: 16} //
)

// Image quad
type Image struct {
	pt  *packTex
	w   int32
	h   int32
	tx0 float32
	tx1 float32
	ty0 float32
	ty1 float32
}

type i32Rect struct{ x, y, w, h int32 }

type ptTree struct {
	r    i32Rect
	s0   *ptTree
	s1   *ptTree
	used bool
}

// pack small pictures into large texture
type packTex struct {
	t ptTree
	x *Res
}

func (t *ptTree) alloc(w, h int32) (ret i32Rect, ok bool) {
	if w <= 0 || h <= 0 {
		return
	}
	//ASSERT(s0 && s1 || !s0 && !s1);

	if t.s0 != nil && t.s1 != nil {
		if ret, ok = t.s0.alloc(w, h); ok {
			return
		}
		return t.s1.alloc(w, h)
	}
	if w > t.r.w || h > t.r.h || t.used {
		ok = false
		return
	}
	if w == t.r.w && h == t.r.h {
		t.used = true
		return t.r, true
	}
	if h == t.r.h {
		ret = t.r
		ret.w = w
		t.r.x += w
		t.r.w -= w
		return ret, true
	}
	if w == t.r.w {
		ret = t.r
		ret.h = h
		t.r.y += h
		t.r.h -= h
		return ret, true
	}
	t.s0 = new(ptTree)
	t.s0.r = t.r
	t.s1 = new(ptTree)
	t.s1.r = t.r
	dx := t.r.w - w
	dy := t.r.h - h
	if dx > dy {
		t.s0.r.x += w
		t.s0.r.w -= w
		t.s1.r.w = w
		t.s1.r.y += h
		t.s1.r.h -= h
	} else {
		t.s0.r.h = h
		t.s0.r.x += w
		t.s0.r.w -= w
		t.s1.r.y += h
		t.s1.r.h -= h
	}
	ret = t.r
	ret.w = w
	ret.h = h
	return ret, true
}

func newPackTex(size int32) *packTex {
	ct := new(packTex)

	ct.t.r.x = 0
	ct.t.r.y = 0
	ct.t.r.w = size
	ct.t.r.h = size

	ct.x = GenTexture("*packTex")

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.BindTexture(gl.TEXTURE_2D, ct.x.ID())
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	pixels := make([]byte, size*size*4)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, size, size, 0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&pixels[0]))
	DbgCheckError()
	return ct
}

// allocate space for picture
func allocPicSpace(w, h int32) (pt *packTex, rect i32Rect) {
	var ok bool
	for _, p := range packTextures {
		if rect, ok = p.t.alloc(w, h); ok {
			pt = p
			return
		}
	}
	p := newPackTex(sizePackTex)
	// if packTextures == nil {
	// 	fallbackPic.width, fallbackPic.height = 16, 16
	// }
	packTextures = append(packTextures, p)
	if rect, ok = p.t.alloc(w, h); ok {
		pt = p
		return
	}
	pt = nil
	return
}
