package color

import (
	"fmt"
	"image/color"
	"math"
	"strings"
)

type Color struct {
	R, G, B, A uint8
}

func (cr Color) String() string {
	if cr.A == 255 {
		return fmt.Sprintf("#%02X%02X%02X", cr.R, cr.G, cr.B)
	}
	return fmt.Sprintf("#%02X%02X%02X%02X", cr.R, cr.G, cr.B, cr.A)
}

func (cr Color) NRGBAf() (r, g, b, a float64) {
	const f = 1.0 / float64(0xFF)
	r = float64(cr.R) * f
	g = float64(cr.G) * f
	b = float64(cr.B) * f
	a = float64(cr.A) * f
	return
}

var ColorModel color.Model = color.ModelFunc(colorModel)

func colorModel(c color.Color) color.Color {
	if _, ok := c.(Color); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	if a == 0xffff {
		return Color{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 0xff}
	}
	if a == 0 {
		return Color{0, 0, 0, 0}
	}
	// Since Color.RGBA returns a alpha-premultiplied color, we should have r <= a && g <= a && b <= a.
	r = (r * 0xffff) / a
	g = (g * 0xffff) / a
	b = (b * 0xffff) / a
	return Color{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R)
	r |= r << 8
	r *= uint32(c.A)
	r /= 0xff
	g = uint32(c.G)
	g |= g << 8
	g *= uint32(c.A)
	g /= 0xff
	b = uint32(c.B)
	b |= b << 8
	b *= uint32(c.A)
	b /= 0xff
	a = uint32(c.A)
	a |= a << 8
	return
}

func Parse(s string) Color {
	if s == "" {
		return Color{R: 0, G: 0, B: 0, A: 255}
	}
	if s[0] == '#' {
		var r, g, b, a uint8
		switch len(s) {
		case 7:
			fmt.Sscanf(s, "#%02X%02X%02X", &r, &g, &b)
			a = 255
		case 9:
			fmt.Sscanf(s, "#%02X%02X%02X%02X", &r, &g, &b, &a)
		case 4:
			fmt.Sscanf(s, "#%01X%01X%01X", &r, &g, &b)
			a = 255
		case 5:
			fmt.Sscanf(s, "#%01X%01X%01X%01X", &r, &g, &b, &a)
		default:
			return Color{R: 0, G: 0, B: 0, A: 255}
		}
		return Color{R: r, G: g, B: b, A: a}
	}

	if namedColors == nil {
		initNamedColors()
	}

	if c, ok := namedColors[strings.ToLower(s)]; ok {
		return c
	}

	if strings.HasSuffix(s, "色") {
		if c, ok := namedColors[s[:len(s)-1]]; ok {
			return c
		}
	}

	return Color{R: 0, G: 0, B: 0, A: 255}
}

type HSL struct {
	H, S, L, A float64
}

func floatToInt32(x float64) uint32 {
	y := uint32(x * float64(0xFFFF))
	if y&0xFFFF0000 == 0 {
		return y
	} else {
		return 0xFFFF
	}
}

func (hsl HSL) RGBA() (r, g, b, a uint32) {
	a = floatToInt32(hsl.A)

	if hsl.S == 0 {
		r = floatToInt32(hsl.L)
		g = r
		b = r
		return
	}
	h := hsl.H
	s := hsl.S
	l := hsl.L

	var t2 float64
	if l < 0.5 {
		t2 = l * (1.0 + s)
	} else {
		t2 = l + s - l*s
	}

	t1 := 2.0*l - t2
	t3_r := h + 1.0/3.0
	t3_g := h
	t3_b := h - 1.0/3.0
	if t3_r > 1.0 {
		t3_r -= 1.0
	}
	if t3_b < 0.0 {
		t3_b += 1.0
	}

	var r1, g1, b1 float64
	if 6.0*t3_r < 1.0 {
		r1 = t1 + (t2-t1)*6.0*t3_r
	} else if 2.0*t3_r < 1.0 {
		r1 = t2
	} else if 3.0*t3_r < 2.0 {
		r1 = t1 + (t2-t1)*6.0*((2.0/3.0)-t3_r)
	} else {
		r1 = t1
	}

	if 6.0*t3_g < 1.0 {
		g1 = t1 + (t2-t1)*6.0*t3_g
	} else if 2.0*t3_g < 1.0 {
		g1 = t2
	} else if 3.0*t3_g < 2.0 {
		g1 = t1 + (t2-t1)*6.0*((2.0/3.0)-t3_g)
	} else {
		g1 = t1
	}

	if 6.0*t3_b < 1.0 {
		b1 = t1 + (t2-t1)*6.0*t3_b
	} else if 2.0*t3_b < 1.0 {
		b1 = t2
	} else if 3.0*t3_b < 2.0 {
		b1 = t1 + (t2-t1)*6.0*((2.0/3.0)-t3_b)
	} else {
		b1 = t1
	}

	r = floatToInt32(r1)
	g = floatToInt32(g1)
	b = floatToInt32(b1)
	return
}

var HSLModel color.Model = color.ModelFunc(hslModel)

func hslModel(c color.Color) color.Color {
	if _, ok := c.(HSL); ok {
		return c
	}

	const f = 1.0 / float64(0xFFFF)
	ri, gi, bi, ai := c.RGBA()
	r1, g1, b1, a1 := float64(ri)*f, float64(gi)*f, float64(bi)*f, float64(ai)*f
	var h, s, l float64

	max1 := r1
	max1 = math.Max(max1, g1)
	max1 = math.Max(max1, b1)
	min1 := r1
	min1 = math.Min(min1, g1)
	min1 = math.Min(min1, b1)

	if max1 == min1 {
		return HSL{0.627, 0, r1, a1}
	}

	max_sub_min := max1 - min1
	max_add_min := max1 + min1
	l = max_add_min / 2.0

	if l < 0.5 {
		s = max_sub_min / max_add_min
	} else {
		s = max_sub_min / (2.0 - max_add_min)
	}

	if r1 == max1 {
		h = (g1 - b1) / max_sub_min
	} else if g1 == max1 {
		h = 2.0 + (b1-r1)/max_sub_min
	} else {
		h = 4.0 + (r1-g1)/max_sub_min
	}

	if h < 0 {
		h += 6.0
	}

	return HSL{h / 6.0, s, l, a1}
}

func FromBgrUint32(c uint32) Color {
	return Color{
		B: uint8(c & 0xff),
		G: uint8((c >> 8) & 0xff),
		R: uint8((c >> 16) & 0xff),
		A: uint8(0xff)}
}

var namedColors map[string]Color

func initNamedColors() {

	namedColors = make(map[string]Color)

	// pinkColors
	namedColors[`pink`] = FromBgrUint32(0xffc0cb)
	namedColors[`lightpink`] = FromBgrUint32(0xffb6c1)
	namedColors[`hotpink`] = FromBgrUint32(0xff69b4)
	namedColors[`deeppink`] = FromBgrUint32(0xff1493)
	namedColors[`palevioletred`] = FromBgrUint32(0xdb7093)
	namedColors[`mediumvioletred`] = FromBgrUint32(0xc71585)

	//redColors
	namedColors[`lightsalmon`] = FromBgrUint32(0xffa07a)
	namedColors[`salmon`] = FromBgrUint32(0xfa8072)
	namedColors[`darksalmon`] = FromBgrUint32(0xe9967a)
	namedColors[`lightcoral`] = FromBgrUint32(0xf08080)
	namedColors[`indianred`] = FromBgrUint32(0xcd5c5c)
	namedColors[`crimson`] = FromBgrUint32(0xdc143c)
	namedColors[`firebrick`] = FromBgrUint32(0xb22222)
	namedColors[`darkred`] = FromBgrUint32(0x8b0000)
	namedColors[`red`] = FromBgrUint32(0xff0000)

	//orangeColors
	namedColors[`orangered`] = FromBgrUint32(0xff4500)
	namedColors[`tomato`] = FromBgrUint32(0xff6347)
	namedColors[`coral`] = FromBgrUint32(0xff7f50)
	namedColors[`darkorange`] = FromBgrUint32(0xff8c00)
	namedColors[`orange`] = FromBgrUint32(0xffa500)

	//yellowColors
	namedColors[`yellow`] = FromBgrUint32(0xffff00)
	namedColors[`lightyellow`] = FromBgrUint32(0xffffe0)
	namedColors[`lemonchiffon`] = FromBgrUint32(0xfffacd)
	namedColors[`lightgoldenrodyellow`] = FromBgrUint32(0xfafad2)
	namedColors[`papayawhip`] = FromBgrUint32(0xffefd5)
	namedColors[`moccasin`] = FromBgrUint32(0xffe4b5)
	namedColors[`peachpuff`] = FromBgrUint32(0xffdab9)
	namedColors[`palegoldenrod`] = FromBgrUint32(0xeee8aa)
	namedColors[`khaki`] = FromBgrUint32(0xf0e68c)
	namedColors[`darkkhaki`] = FromBgrUint32(0xbdb76b)
	namedColors[`gold`] = FromBgrUint32(0xffd700)

	//brownColors
	namedColors[`cornsilk`] = FromBgrUint32(0xfff8dc)
	namedColors[`blanchedalmond`] = FromBgrUint32(0xffebcd)
	namedColors[`bisque`] = FromBgrUint32(0xffe4c4)
	namedColors[`navajowhite`] = FromBgrUint32(0xffdead)
	namedColors[`wheat`] = FromBgrUint32(0xf5deb3)
	namedColors[`burlywood`] = FromBgrUint32(0xdeb887)
	namedColors[`tan`] = FromBgrUint32(0xd2b48c)
	namedColors[`rosybrown`] = FromBgrUint32(0xbc8f8f)
	namedColors[`sandybrown`] = FromBgrUint32(0xf4a460)
	namedColors[`goldenrod`] = FromBgrUint32(0xdaa520)
	namedColors[`darkgoldenrod`] = FromBgrUint32(0xb8860b)
	namedColors[`peru`] = FromBgrUint32(0xcd853f)
	namedColors[`chocolate`] = FromBgrUint32(0xd2691e)
	namedColors[`saddlebrown`] = FromBgrUint32(0x8b4513)
	namedColors[`sienna`] = FromBgrUint32(0xa0522d)
	namedColors[`brown`] = FromBgrUint32(0xa52a2a)
	namedColors[`maroon`] = FromBgrUint32(0x800000)

	//greenColors
	namedColors[`darkolivegreen`] = FromBgrUint32(0x556b2f)
	namedColors[`olive`] = FromBgrUint32(0x808000)
	namedColors[`olivedrab`] = FromBgrUint32(0x6b8e23)
	namedColors[`yellowgreen`] = FromBgrUint32(0x9acd32)
	namedColors[`limegreen`] = FromBgrUint32(0x32cd32)
	namedColors[`lime`] = FromBgrUint32(0x00ff00)
	namedColors[`lawngreen`] = FromBgrUint32(0x7cfc00)
	namedColors[`chartreuse`] = FromBgrUint32(0x7fff00)
	namedColors[`greenyellow`] = FromBgrUint32(0xadff2f)
	namedColors[`springgreen`] = FromBgrUint32(0x00ff7f)
	namedColors[`mediumspringgreen`] = FromBgrUint32(0x00fa9a)
	namedColors[`lightgreen`] = FromBgrUint32(0x90ee90)
	namedColors[`palegreen`] = FromBgrUint32(0x98fb98)
	namedColors[`darkseagreen`] = FromBgrUint32(0x8fbc8f)
	namedColors[`mediumseagreen`] = FromBgrUint32(0x3cb371)
	namedColors[`seagreen`] = FromBgrUint32(0x2e8b57)
	namedColors[`forestgreen`] = FromBgrUint32(0x228b22)
	namedColors[`green`] = FromBgrUint32(0x008000)
	namedColors[`darkgreen`] = FromBgrUint32(0x006400)

	//cyanColors
	namedColors[`mediumaquamarine`] = FromBgrUint32(0x66cdaa)
	namedColors[`aqua`] = FromBgrUint32(0x00ffff)
	namedColors[`cyan`] = FromBgrUint32(0x00ffff)
	namedColors[`lightcyan`] = FromBgrUint32(0xe0ffff)
	namedColors[`paleturquoise`] = FromBgrUint32(0xafeeee)
	namedColors[`aquamarine`] = FromBgrUint32(0x7fffd4)
	namedColors[`turquoise`] = FromBgrUint32(0x40e0d0)
	namedColors[`mediumturquoise`] = FromBgrUint32(0x48d1cc)
	namedColors[`darkturquoise`] = FromBgrUint32(0x00ced1)
	namedColors[`lightseagreen`] = FromBgrUint32(0x20b2aa)
	namedColors[`cadetblue`] = FromBgrUint32(0x5f9ea0)
	namedColors[`darkcyan`] = FromBgrUint32(0x008b8b)
	namedColors[`teal`] = FromBgrUint32(0x008080)

	//blueColors
	namedColors[`lightsteelblue`] = FromBgrUint32(0xb0c4de)
	namedColors[`powderblue`] = FromBgrUint32(0xb0e0e6)
	namedColors[`lightblue`] = FromBgrUint32(0xadd8e6)
	namedColors[`skyblue`] = FromBgrUint32(0x87ceeb)
	namedColors[`lightskyblue`] = FromBgrUint32(0x87cefa)
	namedColors[`deepskyblue`] = FromBgrUint32(0x00bfff)
	namedColors[`dodgerblue`] = FromBgrUint32(0x1e90ff)
	namedColors[`cornflowerblue`] = FromBgrUint32(0x6495ed)
	namedColors[`steelblue`] = FromBgrUint32(0x4682b4)
	namedColors[`royalblue`] = FromBgrUint32(0x4169e1)
	namedColors[`blue`] = FromBgrUint32(0x0000ff)
	namedColors[`mediumblue`] = FromBgrUint32(0x0000cd)
	namedColors[`darkblue`] = FromBgrUint32(0x00008b)
	namedColors[`navy`] = FromBgrUint32(0x000080)
	namedColors[`midnightblue`] = FromBgrUint32(0x191970)

	// 中文颜色
	namedColors[`黑`] = FromBgrUint32(0x000000)
	namedColors[`昏灰`] = FromBgrUint32(0x696969)
	namedColors[`灰`] = FromBgrUint32(0x808080)
	namedColors[`暗灰`] = FromBgrUint32(0xA9A9A9)
	namedColors[`银`] = FromBgrUint32(0xC0C0C0)
	namedColors[`亮灰`] = FromBgrUint32(0xD3D3D3)
	namedColors[`庚斯博罗灰`] = FromBgrUint32(0xDCDCDC)
	namedColors[`白烟`] = FromBgrUint32(0xF5F5F5)
	namedColors[`白`] = FromBgrUint32(0xFFFFFF)
	namedColors[`雪`] = FromBgrUint32(0xFFFAFA)
	namedColors[`铁灰`] = FromBgrUint32(0x625B57)
	namedColors[`沙棕`] = FromBgrUint32(0xE6C3C3)
	namedColors[`玫瑰褐`] = FromBgrUint32(0xBC8F8F)
	namedColors[`亮珊瑚`] = FromBgrUint32(0xF08080)
	namedColors[`印度红`] = FromBgrUint32(0xCD5C5C)
	namedColors[`褐`] = FromBgrUint32(0xA52A2A)
	namedColors[`耐火砖红`] = FromBgrUint32(0xB22222)
	namedColors[`栗`] = FromBgrUint32(0x800000)
	namedColors[`暗红`] = FromBgrUint32(0x8B0000)
	namedColors[`鲜红`] = FromBgrUint32(0xE60000)
	namedColors[`红`] = FromBgrUint32(0xFF0000)
	namedColors[`柿子橙`] = FromBgrUint32(0xFF4D40)
	namedColors[`雾玫瑰`] = FromBgrUint32(0xFFE4E1)
	namedColors[`鲑红`] = FromBgrUint32(0xFA8072)
	namedColors[`腥红`] = FromBgrUint32(0xFF2400)
	namedColors[`蕃茄红`] = FromBgrUint32(0xFF6347)
	namedColors[`暗鲑红`] = FromBgrUint32(0xE9967A)
	namedColors[`珊瑚红`] = FromBgrUint32(0xFF7F50)
	namedColors[`橙红`] = FromBgrUint32(0xFF4500)
	namedColors[`亮鲑红`] = FromBgrUint32(0xFFA07A)
	namedColors[`朱红`] = FromBgrUint32(0xFF4D00)
	namedColors[`赭黄`] = FromBgrUint32(0xA0522D)
	namedColors[`热带橙`] = FromBgrUint32(0xFF8033)
	namedColors[`驼`] = FromBgrUint32(0xA16B47)
	namedColors[`杏黄`] = FromBgrUint32(0xE69966)
	namedColors[`椰褐`] = FromBgrUint32(0x4D1F00)
	namedColors[`海贝`] = FromBgrUint32(0xFFF5EE)
	namedColors[`鞍褐`] = FromBgrUint32(0x8B4513)
	namedColors[`巧克力`] = FromBgrUint32(0xD2691E)
	namedColors[`燃橙`] = FromBgrUint32(0xCC5500)
	namedColors[`阳橙`] = FromBgrUint32(0xFF7300)
	namedColors[`粉扑桃`] = FromBgrUint32(0xFFDAB9)
	namedColors[`沙褐`] = FromBgrUint32(0xF4A460)
	namedColors[`古铜`] = FromBgrUint32(0xB87333)
	namedColors[`亚麻`] = FromBgrUint32(0xFAF0E6)
	namedColors[`蜜橙`] = FromBgrUint32(0xFFB366)
	namedColors[`秘鲁`] = FromBgrUint32(0xCD853F)
	namedColors[`乌贼墨`] = FromBgrUint32(0x704214)
	namedColors[`赭`] = FromBgrUint32(0xCC7722)
	namedColors[`陶坯黄`] = FromBgrUint32(0xFFE4C4)
	namedColors[`橘`] = FromBgrUint32(0xF28500)
	namedColors[`暗橙`] = FromBgrUint32(0xFF8C00)
	namedColors[`古董白`] = FromBgrUint32(0xFAEBD7)
	namedColors[`日晒`] = FromBgrUint32(0xD2B48C)
	namedColors[`硬木`] = FromBgrUint32(0xDEB887)
	namedColors[`杏仁白`] = FromBgrUint32(0xFFEBCD)
	namedColors[`那瓦霍白`] = FromBgrUint32(0xFFDEAD)
	namedColors[`万寿菊黄`] = FromBgrUint32(0xFF9900)
	namedColors[`蕃木瓜`] = FromBgrUint32(0xFFEFD5)
	namedColors[`灰土`] = FromBgrUint32(0xCCB38C)
	namedColors[`卡其`] = FromBgrUint32(0x996B1F)
	namedColors[`鹿皮鞋`] = FromBgrUint32(0xFFE4B5)
	namedColors[`旧蕾丝`] = FromBgrUint32(0xFDF5E6)
	namedColors[`小麦`] = FromBgrUint32(0xF5DEB3)
	namedColors[`桃`] = FromBgrUint32(0xFFE5B4)
	namedColors[`橙`] = FromBgrUint32(0xFFA500)
	namedColors[`花卉白`] = FromBgrUint32(0xFFFAF0)
	namedColors[`金菊`] = FromBgrUint32(0xDAA520)
	namedColors[`暗金菊`] = FromBgrUint32(0xB8860B)
	namedColors[`咖啡`] = FromBgrUint32(0x4D3900)
	namedColors[`茉莉黄`] = FromBgrUint32(0xE6C35C)
	namedColors[`琥珀`] = FromBgrUint32(0xFFBF00)
	namedColors[`玉米丝`] = FromBgrUint32(0xFFF8DC)
	namedColors[`铬黄`] = FromBgrUint32(0xE6B800)
	namedColors[`金`] = FromBgrUint32(0xFFD700)
	namedColors[`柠檬绸`] = FromBgrUint32(0xFFFACD)
	namedColors[`亮卡其`] = FromBgrUint32(0xF0E68C)
	namedColors[`灰金菊`] = FromBgrUint32(0xEEE8AA)
	namedColors[`暗卡其`] = FromBgrUint32(0xBDB76B)
	namedColors[`含羞草黄`] = FromBgrUint32(0xE6D933)
	namedColors[`奶油`] = FromBgrUint32(0xFFFDD0)
	namedColors[`象牙`] = FromBgrUint32(0xFFFFF0)
	namedColors[`米黄`] = FromBgrUint32(0xF5F5DC)
	namedColors[`亮黄`] = FromBgrUint32(0xFFFFE0)
	namedColors[`亮金菊黄`] = FromBgrUint32(0xFAFAD2)
	namedColors[`香槟黄`] = FromBgrUint32(0xFFFF99)
	namedColors[`芥末黄`] = FromBgrUint32(0xCCCC4D)
	namedColors[`月黄`] = FromBgrUint32(0xFFFF4D)
	namedColors[`橄榄`] = FromBgrUint32(0x808000)
	namedColors[`鲜黄`] = FromBgrUint32(0xFFFF00)
	namedColors[`黄`] = FromBgrUint32(0xFFFF00)
	namedColors[`苔藓绿`] = FromBgrUint32(0x697723)
	namedColors[`亮柠檬绿`] = FromBgrUint32(0xCCFF00)
	namedColors[`橄榄军服绿`] = FromBgrUint32(0x6B8E23)
	namedColors[`黄绿`] = FromBgrUint32(0x9ACD32)
	namedColors[`暗橄榄绿`] = FromBgrUint32(0x556B2F)
	namedColors[`苹果绿`] = FromBgrUint32(0x8CE600)
	namedColors[`绿黄`] = FromBgrUint32(0xADFF2F)
	namedColors[`草绿`] = FromBgrUint32(0x99E64D)
	namedColors[`草坪绿`] = FromBgrUint32(0x7CFC00)
	namedColors[`查特酒绿`] = FromBgrUint32(0x7FFF00)
	namedColors[`叶绿`] = FromBgrUint32(0x73B839)
	namedColors[`嫩绿`] = FromBgrUint32(0x99FF4D)
	namedColors[`明绿`] = FromBgrUint32(0x66FF00)
	namedColors[`钴绿`] = FromBgrUint32(0x66FF59)
	namedColors[`蜜瓜绿`] = FromBgrUint32(0xF0FFF0)
	namedColors[`暗海绿`] = FromBgrUint32(0x8FBC8F)
	namedColors[`亮绿`] = FromBgrUint32(0x90EE90)
	namedColors[`灰绿`] = FromBgrUint32(0x98FB98)
	namedColors[`常春藤绿`] = FromBgrUint32(0x36BF36)
	namedColors[`森林绿`] = FromBgrUint32(0x228B22)
	namedColors[`柠檬绿`] = FromBgrUint32(0x32CD32)
	namedColors[`暗绿`] = FromBgrUint32(0x006400)
	namedColors[`绿`] = FromBgrUint32(0x008000)
	namedColors[`鲜绿`] = FromBgrUint32(0x00FF00)
	namedColors[`孔雀石绿`] = FromBgrUint32(0x22C32E)
	namedColors[`薄荷绿`] = FromBgrUint32(0x16982B)
	namedColors[`青瓷绿`] = FromBgrUint32(0x73E68C)
	namedColors[`碧绿`] = FromBgrUint32(0x50C878)
	namedColors[`绿松石绿`] = FromBgrUint32(0x4DE680)
	namedColors[`铬绿`] = FromBgrUint32(0x127436)
	namedColors[`苍`] = FromBgrUint32(0xA6FFCC)
	namedColors[`海绿`] = FromBgrUint32(0x2E8B57)
	namedColors[`中海绿`] = FromBgrUint32(0x3CB371)
	namedColors[`薄荷奶油`] = FromBgrUint32(0xF5FFFA)
	namedColors[`春绿`] = FromBgrUint32(0x00FF80)
	namedColors[`孔雀绿`] = FromBgrUint32(0x00A15C)
	namedColors[`中春绿`] = FromBgrUint32(0x00FA9A)
	namedColors[`中碧蓝`] = FromBgrUint32(0x66CDAA)
	namedColors[`碧蓝`] = FromBgrUint32(0x7FFFD4)
	namedColors[`青蓝`] = FromBgrUint32(0x0DBF8C)
	namedColors[`水蓝`] = FromBgrUint32(0x66FFE6)
	namedColors[`绿松石蓝`] = FromBgrUint32(0x33E6CC)
	namedColors[`绿松石`] = FromBgrUint32(0x30D5C8)
	namedColors[`亮海绿`] = FromBgrUint32(0x20B2AA)
	namedColors[`中绿松石`] = FromBgrUint32(0x48D1CC)
	namedColors[`亮青`] = FromBgrUint32(0xE0FFFF)
	namedColors[`浅蓝`] = FromBgrUint32(0xE0FFFF)
	namedColors[`灰绿松石`] = FromBgrUint32(0xAFEEEE)
	namedColors[`暗岩灰`] = FromBgrUint32(0x2F4F4F)
	namedColors[`凫绿`] = FromBgrUint32(0x008080)
	namedColors[`暗青`] = FromBgrUint32(0x008B8B)
	namedColors[`青`] = FromBgrUint32(0x00FFFF)
	namedColors[`水`] = FromBgrUint32(0xAFDFE4)
	namedColors[`暗绿松石`] = FromBgrUint32(0x00CED1)
	namedColors[`军服蓝`] = FromBgrUint32(0x5F9EA0)
	namedColors[`孔雀蓝`] = FromBgrUint32(0x00808C)
	namedColors[`婴儿粉蓝`] = FromBgrUint32(0xB0E0E6)
	namedColors[`浓蓝`] = FromBgrUint32(0x006374)
	namedColors[`亮蓝`] = FromBgrUint32(0xADD8E6)
	namedColors[`灰蓝`] = FromBgrUint32(0x7AB8CC)
	namedColors[`萨克斯蓝`] = FromBgrUint32(0x4798B3)
	namedColors[`深天蓝`] = FromBgrUint32(0x00BFFF)
	namedColors[`天蓝`] = FromBgrUint32(0x87CEEB)
	namedColors[`亮天蓝`] = FromBgrUint32(0x87CEFA)
	namedColors[`水手蓝`] = FromBgrUint32(0x00477D)
	namedColors[`普鲁士蓝`] = FromBgrUint32(0x003153)
	namedColors[`钢青`] = FromBgrUint32(0x4682B4)
	namedColors[`爱丽丝蓝`] = FromBgrUint32(0xF0F8FF)
	namedColors[`岩灰`] = FromBgrUint32(0x708090)
	namedColors[`亮岩灰`] = FromBgrUint32(0x778899)
	namedColors[`道奇蓝`] = FromBgrUint32(0x1E90FF)
	namedColors[`矿蓝`] = FromBgrUint32(0x004D99)
	namedColors[`湛蓝`] = FromBgrUint32(0x007FFF)
	namedColors[`韦奇伍德瓷蓝`] = FromBgrUint32(0x5686BF)
	namedColors[`亮钢蓝`] = FromBgrUint32(0xB0C4DE)
	namedColors[`钴蓝`] = FromBgrUint32(0x0047AB)
	namedColors[`灰丁宁蓝`] = FromBgrUint32(0x5E86C1)
	namedColors[`矢车菊蓝`] = FromBgrUint32(0x6495ED)
	namedColors[`鼠尾草蓝`] = FromBgrUint32(0x4D80E6)
	namedColors[`暗婴儿粉蓝`] = FromBgrUint32(0x003399)
	namedColors[`蓝宝石`] = FromBgrUint32(0x082567)
	namedColors[`国际奇连蓝`] = FromBgrUint32(0x002FA7)
	namedColors[`蔚蓝`] = FromBgrUint32(0x2A52BE)
	namedColors[`品蓝`] = FromBgrUint32(0x4169E1)
	namedColors[`暗矿蓝`] = FromBgrUint32(0x24367D)
	namedColors[`极浓海蓝`] = FromBgrUint32(0x0033FF)
	namedColors[`天青石蓝`] = FromBgrUint32(0x0D33FF)
	namedColors[`幽灵白`] = FromBgrUint32(0xF8F8FF)
	namedColors[`薰衣草紫`] = FromBgrUint32(0xE6E6FA)
	namedColors[`长春花`] = FromBgrUint32(0xCCCCFF)
	namedColors[`午夜蓝`] = FromBgrUint32(0x191970)
	namedColors[`藏青`] = FromBgrUint32(0x000080)
	namedColors[`暗蓝`] = FromBgrUint32(0x00008B)
	namedColors[`中蓝`] = FromBgrUint32(0x0000CD)
	namedColors[`蓝`] = FromBgrUint32(0x0000FF)
	namedColors[`紫藤`] = FromBgrUint32(0x5C50E6)
	namedColors[`暗岩蓝`] = FromBgrUint32(0x483D8B)
	namedColors[`岩蓝`] = FromBgrUint32(0x6A5ACD)
	namedColors[`中岩蓝`] = FromBgrUint32(0x7B68EE)
	namedColors[`木槿紫`] = FromBgrUint32(0x6640FF)
	namedColors[`紫丁香`] = FromBgrUint32(0xB399FF)
	namedColors[`中紫红`] = FromBgrUint32(0x9370DB)
	namedColors[`紫水晶`] = FromBgrUint32(0x6633CC)
	namedColors[`浅灰紫红`] = FromBgrUint32(0x8674A1)
	namedColors[`缬草紫`] = FromBgrUint32(0x5000B8)
	namedColors[`矿紫`] = FromBgrUint32(0xB8A1CF)
	namedColors[`蓝紫`] = FromBgrUint32(0x8A2BE2)
	namedColors[`紫罗兰`] = FromBgrUint32(0x8B00FF)
	namedColors[`靛`] = FromBgrUint32(0x4B0080)
	namedColors[`暗兰紫`] = FromBgrUint32(0x9932CC)
	namedColors[`暗紫`] = FromBgrUint32(0x9400D3)
	namedColors[`三堇紫`] = FromBgrUint32(0x7400A1)
	namedColors[`锦葵紫`] = FromBgrUint32(0xD94DFF)
	namedColors[`优品紫红`] = FromBgrUint32(0xE680FF)
	namedColors[`中兰紫`] = FromBgrUint32(0xBA55D3)
	namedColors[`淡紫丁香`] = FromBgrUint32(0xE6CFE6)
	namedColors[`蓟紫`] = FromBgrUint32(0xD8BFD8)
	namedColors[`铁线莲紫`] = FromBgrUint32(0xCCA3CC)
	namedColors[`梅红`] = FromBgrUint32(0xDDA0DD)
	namedColors[`亮紫`] = FromBgrUint32(0xEE82EE)
	namedColors[`紫`] = FromBgrUint32(0x800080)
	namedColors[`暗洋红`] = FromBgrUint32(0x8B008B)
	namedColors[`洋红`] = FromBgrUint32(0xFF00FF)
	namedColors[`品红`] = FromBgrUint32(0xF400A1)
	namedColors[`兰紫`] = FromBgrUint32(0xDA70D6)
	namedColors[`浅珍珠红`] = FromBgrUint32(0xFFB3E6)
	namedColors[`陈玫红`] = FromBgrUint32(0xB85798)
	namedColors[`浅玫瑰红`] = FromBgrUint32(0xFF66CC)
	namedColors[`中青紫红`] = FromBgrUint32(0xC71585)
	namedColors[`洋玫瑰红`] = FromBgrUint32(0xFF0DA6)
	namedColors[`玫瑰红`] = FromBgrUint32(0xFF007F)
	namedColors[`红宝石`] = FromBgrUint32(0xCC0080)
	namedColors[`山茶红`] = FromBgrUint32(0xE63995)
	namedColors[`深粉红`] = FromBgrUint32(0xFF1493)
	namedColors[`火鹤红`] = FromBgrUint32(0xE68AB8)
	namedColors[`浅珊瑚红`] = FromBgrUint32(0xFF80BF)
	namedColors[`暖粉红`] = FromBgrUint32(0xFF69B4)
	namedColors[`勃艮第酒红`] = FromBgrUint32(0x470024)
	namedColors[`尖晶石红`] = FromBgrUint32(0xFF73B3)
	namedColors[`胭脂红`] = FromBgrUint32(0xE6005C)
	namedColors[`浅粉红`] = FromBgrUint32(0xFFD9E6)
	namedColors[`枢机红`] = FromBgrUint32(0x990036)
	namedColors[`薰衣草紫红`] = FromBgrUint32(0xFFF0F5)
	namedColors[`灰紫红`] = FromBgrUint32(0xDB7093)
	namedColors[`樱桃红`] = FromBgrUint32(0xDE3163)
	namedColors[`浅鲑红`] = FromBgrUint32(0xFF8099)
	namedColors[`绯红`] = FromBgrUint32(0xDC143C)
	namedColors[`粉红`] = FromBgrUint32(0xFFC0CB)
	namedColors[`亮粉红`] = FromBgrUint32(0xFFB6C1)
	namedColors[`壳黄红`] = FromBgrUint32(0xFFB3BF)
	namedColors[`茜红`] = FromBgrUint32(0xE32636)

}
