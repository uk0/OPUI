package glman

import libColor "tetra/lib/color"
import imageColor "image/color"

// Color is floating point color
type Color [4]float32

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the color. Each value ranges within [0, 0xffff], but is represented
// by a uint32 so that multiplying by a blend factor up to 0xffff will not
// overflow.
//
// An alpha-premultiplied color component c has been scaled by alpha (a),
// so has valid values 0 <= c <= a.
//
// implement image/color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(uint16(c[0] * 65525 * c[3]))
	g = uint32(uint16(c[1] * 65525 * c[3]))
	b = uint32(uint16(c[2] * 65525 * c[3]))
	a = uint32(uint16(c[3] * 65525))
	return
}

// SetBytes set color from r, g, b, a components
func (c *Color) SetBytes(r, g, b, a byte) {
	(*c)[0] = float32(r) / 255
	(*c)[1] = float32(g) / 255
	(*c)[2] = float32(b) / 255
	(*c)[3] = float32(a) / 255
}

// Copy from a image/color.Color
func (c *Color) Copy(s imageColor.Color) {
	switch x := s.(type) {
	case Color:
		*c = x
	case libColor.Color:
		c.SetBytes(x.R, x.G, x.B, x.A)
	default:
		r, g, b, a := s.RGBA()
		f := 1 / float32(a)
		(*c)[0] = float32(r) * f
		(*c)[1] = float32(g) * f
		(*c)[2] = float32(b) * f
		(*c)[3] = float32(a) / float32(65535)
	}
}

// Parse color form string
func (c *Color) Parse(s string) {
	c.Copy(libColor.Parse(s))
}

// MkRGBAF make a Color form r,g,b,a component
func MkRGBAF(r, g, b, a byte) (c Color) {
	c.SetBytes(r, g, b, a)
	return
}
