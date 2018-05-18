// Package wfobj provide interface to load Wave Front 3d models and materials
package wfobj

import (
	"fmt"
	"io"
	"log"
	"strconv"
	scan "text/scanner"
)

// GroupData is vertex group
type GroupData struct {
	Name     string
	Material string
	SzPos    int       // number components of position per vertex
	SzTC     int       // number components of texcoord per vertex
	SzNorm   int       // number components of nromal per vertex
	V        []float32 // interlacing position, texcoord, normal
	S        int
}

func (g GroupData) String() string {
	return fmt.Sprintf("[group \"%s\": [%d]pos,[%d]tc,[%d]norm v=%d mtl=\"%s\"]",
		g.Name, g.SzPos, g.SzTC, g.SzNorm, len(g.V)/(g.SzPos+g.SzTC+g.SzNorm), g.Material)
}

// ModelData is Wave Front MTL 3d model, not loaded into GPU
type ModelData struct {
	Name    string
	MtlLibs []string
	Groups  []*GroupData
}

func (m ModelData) String() string {
	return fmt.Sprintf("[wfobj \"%s\" MtlLibs=%v %v]", m.Name, m.MtlLibs, m.Groups)
}

// ParseObj parse obj format model file.
func ParseObj(r io.Reader, filename string) (m *ModelData, err error) {
	// a obj file looks like this:
	// -------------
	// # commemts
	// mtllib Rock1.mtl
	// o Plane
	// v 6.083834 0.000000 6.083834
	// v -6.083834 0.000000 6.083834
	// v 6.083834 0.000000 -6.083834
	// v -6.083834 0.000000 -6.083834
	// usemtl Material.001
	// s off
	// f 2 1 3 4
	// o Cube
	// v 0.896930 -0.116701 -1.078061
	// v 0.736314 -0.076033 1.066762
	// v -1.052088 -0.064600 0.954513
	// vt 0.137636 0.563411
	// vt 0.058285 0.569867
	// vt 0.082613 0.506661
	// usemtl Material
	// s 1
	// f 79/1 34/2 7/3 48/4
	// f 80/5 54/6 11/7 41/8
	// --------------

	m = new(ModelData)
	// we alloc a dummy material to avoid boring m==nil check, values before first
	// "newmtl" will strore into it, but just discard silently
	g := new(GroupData)
	s := new(scan.Scanner).Init(r)
	s.Filename = filename
	tok := s.Scan()

	tmpPos := make([][3]float32, 0, 4096)
	tmpTC := make([][3]float32, 0, 4096)
	tmpNorm := make([][3]float32, 0, 4096)
	var x [3]float32
	var n int
	var t string
	//var mtl string
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
		case "mtllib":
			if t, err = readString(s); err != nil {
				return nil, err
			}
			m.MtlLibs = append(m.MtlLibs, t)
		case "usemtl":
			if t, err = readString(s); err != nil {
				return nil, err
			}
			g.Material = t
		case "o", "g":
			if t, err = readString(s); err != nil {
				return nil, err
			}
			//tmpPos = tmpPos[:0]
			//tmpTC = tmpTC[:0]
			//tmpNorm = tmpNorm[:0]
			g = new(GroupData)
			g.Name = t
			m.Groups = append(m.Groups, g)
		case "s":
			if t, err = readString(s); err != nil {
				return nil, err
			}
			x, _ := strconv.ParseInt(t, 10, 32)
			g.S = int(x)
		case "v":
			if x, n = readFloats(s); n == 0 {
				return nil, (*errPosition)(&s.Position)
			}
			tmpPos = append(tmpPos, x)
			if n > g.SzPos {
				g.SzPos = n
				if n > 3 {
					panic("n > 3")
				}
			}
		case "vt":
			if x, n = readFloats(s); n == 0 {
				return nil, (*errPosition)(&s.Position)
			}
			tmpTC = append(tmpTC, x)
			if n > g.SzTC {
				g.SzTC = n
			}
		case "vn":
			if x, n = readFloats(s); n == 0 {
				return nil, (*errPosition)(&s.Position)
			}
			tmpNorm = append(tmpNorm, x)
			if n > g.SzNorm {
				g.SzNorm = n
			}
		case "f":
			// f 79/1/1   34/2/3    7/3/7    48/4/5/6
			//    p t n    p t n    p t n    p  t  n
			ln := s.Position.Line
			tok = s.Scan()
			if tok == scan.EOF || s.Position.Line != ln {
				return nil, fmt.Errorf("%v, empty \"f\" line", (*errPosition)(&s.Position))
			}
			// for each vertex, until line end
			for {

				// positon[/texcoord][/normal]

				var y int64
				var i int

				// position
				if y, err = strconv.ParseInt(s.TokenText(), 10, 32); err != nil {
					return nil, (*errPosition)(&s.Position)
				}
				i = int(y) - 1
				if i < 0 || i >= len(tmpPos) {
					return nil, fmt.Errorf("%v, position index overflow: i=%d, len(pos)=%d",
						(*errPosition)(&s.Position), i, len(tmpPos))
				}
				for k := 0; k < g.SzPos; k++ {
					g.V = append(g.V, tmpPos[i][k])
				}

				// texcoord
				tok = s.Scan()
				if tok == scan.EOF || s.Position.Line != ln {
					break
				}
				if s.TokenText() != "/" {
					continue
				}
				tok = s.Scan()
				if tok == scan.EOF || s.Position.Line != ln {
					break
				}
				if y, err = strconv.ParseInt(s.TokenText(), 10, 32); err != nil {
					return nil, (*errPosition)(&s.Position)
				}
				i = int(y) - 1
				if i < 0 || i >= len(tmpTC) {
					return nil, fmt.Errorf("%v, texcoord index overflow", (*errPosition)(&s.Position))
				}
				for k := 0; k < g.SzTC; k++ {
					g.V = append(g.V, tmpTC[i][k])
				}
				tok = s.Scan()
				if tok == scan.EOF || s.Position.Line != ln {
					break
				}
				if s.TokenText() != "/" {
					continue
				}

				// normal
				tok = s.Scan()
				if tok == scan.EOF || s.Position.Line != ln {
					break
				}
				if y, err = strconv.ParseInt(s.TokenText(), 10, 32); err != nil {
					return nil, (*errPosition)(&s.Position)
				}
				i = int(y) - 1
				if i < 0 || i >= len(tmpNorm) {
					return nil, fmt.Errorf("%v, normal index overflow", (*errPosition)(&s.Position))
				}
				for k := 0; k < g.SzNorm; k++ {
					g.V = append(g.V, tmpNorm[i][k])
				}
				tok = s.Scan()
				if tok == scan.EOF || s.Position.Line != ln {
					break
				}
				// there must not extra components
				if s.TokenText() == "/" {
					return nil, fmt.Errorf("%v, extra component after positon/texcoord/normal", (*errPosition)(&s.Position))
				}
			}
			if ln == s.Position.Line {
				discardLine(s)
			}
		default:
			log.Printf("warning: unkown obj line \"%s\": %v", key, (*errPosition)(&s.Position))
			if err = discardLine(s); err != nil {
				return
			}
		}
	}
	return
}
