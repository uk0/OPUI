package wfobj

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"tetra/lib/glman"
	scan "text/scanner"
)

type errPosition scan.Position

func (e *errPosition) Error() string {
	return fmt.Sprintf("%s:%d:%d: parsing error", e.Filename, e.Line, e.Column)
}

func discardLine(s *scan.Scanner) error {
	_, err := readString(s)
	return err
}

// read line as string
func readString(s *scan.Scanner) (str string, err error) {
	ln := s.Position.Line
	tok := s.Scan()
	for tok != scan.EOF && s.Position.Line == ln {
		str += s.TokenText()
		tok = s.Scan()
	}
	str = strings.TrimSpace(str)
	return
}

func readFloat(s *scan.Scanner) (x float32, err error) {
	y, n := readFloats(s)
	if n < 1 {
		return 0, (*errPosition)(&s.Position)
	}
	x = y[0]
	return
}

func readUint32(s *scan.Scanner) (x uint32, err error) {
	ln := s.Position.Line
	tok := s.Scan()
	if tok == scan.EOF || s.Position.Line != ln {
		return 0, (*errPosition)(&s.Position)
	}
	var y uint64
	if y, err = strconv.ParseUint(s.TokenText(), 10, 32); err != nil {
		return
	}
	x = uint32(y)
	discardLine(s)
	return
}

func read3Float(s *scan.Scanner) (x [3]float32, err error) {
	x, n := readFloats(s)
	if n < 3 {
		return x, (*errPosition)(&s.Position)
	}
	return
}

func readFloats(s *scan.Scanner) (x [3]float32, n int) {
	ln := s.Position.Line
	for n = 0; n < 3; n++ {
		tok := s.Scan()
		if tok == scan.EOF || s.Position.Line != ln {
			return
		}
		var y float64
		var err error
		var neg bool
		if s.TokenText() == "-" {
			neg = true
			tok := s.Scan()
			if tok == scan.EOF || s.Position.Line != ln {
				return
			}
		}
		if y, err = strconv.ParseFloat(s.TokenText(), 32); err != nil {
			return
		}
		x[n] = float32(y)
		if neg {
			x[n] = -x[n]
		}
	}
	if s.Position.Line == ln {
		discardLine(s)
	}
	return
}

// ParseMtl parse mtl format material file.
func ParseMtl(r io.Reader, filename string) (ms map[string]*glman.MtlInfo, err error) {

	// a mtl file looks like this:
	// -------------
	// # comment
	// newmtl Material Name
	// Ns 13.725490
	// Ka 0.000000 0.000000 0.000000
	// Kd 0.640000 0.640000 0.640000
	// Ks 0.007937 0.007937 0.007937
	// Ni 1.000000
	// d 1.000000
	// illum 2
	// map_Disp foo.jpg
	// map_Kd bar.jpg
	// --------------

	ms = make(map[string]*glman.MtlInfo)
	// we alloc a dummy material to avoid boring m==nil check, values before first
	// "newmtl" will strore into it, but just discard silently
	m := new(glman.MtlInfo)
	s := new(scan.Scanner).Init(r)
	s.Filename = filename
	tok := s.Scan()
	//ln := s.Position.Line
	var t string
	for tok != scan.EOF {
		// read first token of line
		key := s.TokenText()
		switch key {
		case "":
			tok = s.Scan()
		case "#":
			if err = discardLine(s); err != nil {
				return
			}
		case "newmtl":
			if t, err = readString(s); err != nil {
				return nil, err
			}
			m = new(glman.MtlInfo)
			ms[t] = m
		case "Ns":
			if m.SpecularPower, err = readFloat(s); err != nil {
				return nil, err
			}
		case "Kd":
			if m.Diffuse, err = read3Float(s); err != nil {
				return nil, err
			}
		case "Ka":
			if m.Ambient, err = read3Float(s); err != nil {
				return nil, err
			}
		case "Ks":
			if m.Specular, err = read3Float(s); err != nil {
				return nil, err
			}
		case "d":
			if m.Alpha, err = readFloat(s); err != nil {
				return nil, err
			}
		case "Tr":
			if m.Alpha, err = readFloat(s); err != nil {
				return nil, err
			}
			m.Alpha = 1 - m.Alpha
		case "illum":
			if m.RenderMode, err = readUint32(s); err != nil {
				return nil, err
			}
			//0. Color on and Ambient off
			//1. Color on and Ambient on
			//2. Highlight on
			//3. Reflection on and Ray hx_trace on
			//4. Transparency: Glass on, Reflection: Ray hx_trace on
			//5. Reflection: Fresnel on and Ray hx_trace on
			//6. Transparency: Refraction on, Reflection: Fresnel off and Ray hx_trace on
			//7. Transparency: Refraction on, Reflection: Fresnel on and Ray hx_trace on
			//8. Reflection on and Ray hx_trace off
			//9. Transparency: Glass on, Reflection: Ray hx_trace off
			//10. Casts shadows onto invisible surfaces
		case "map_Kd", "map_Disp":
			if m.MapDiffuse, err = readString(s); err != nil {
				return nil, err
			}
		case "map_Ks":
			if m.MapSpecular, err = readString(s); err != nil {
				return nil, err
			}
		case "map_d":
			if m.MapAlpha, err = readString(s); err != nil {
				return nil, err
			}
		case "map_Bump":
			if m.MapBump, err = readString(s); err != nil {
				return nil, err
			}
		default:
			log.Printf("warning: unkown mtl line \"%s\": %v", key, (*errPosition)(&s.Position))
			if err = discardLine(s); err != nil {
				return
			}
		}
	}
	return
}
