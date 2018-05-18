package glman

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"tetra/lib/store"

	"tetra/internal/gl"
)

var (
	scache = make(map[string]uint32)
	pcache = make(map[string]*Program)

	exttype = map[string]uint32{
		".vert": gl.VERTEX_SHADER,
		".frag": gl.FRAGMENT_SHADER,
	}
	glslVersion = "#version 120\n"
)

var glslIncVert = `
#ifndef GL_ES
#  define precision
#  define mediump
#endif

#if __VERSION__ <= 120 // TODO: which version?
#  define in attribute
#  define out varying
#endif

`

var glslIncFrag = `
#ifndef GL_ES
#  define precision
#  define mediump
#endif

#if __VERSION__ <= 120 // TODO: which version?
#  define in varying
#endif

`

// Program wrap shader program
type Program struct {

	// ID of program
	ID uint32

	// attribute location for vertesx shader
	AttPos   int32 `attrib:"attPos"`
	AttTC    int32 `attrib:"attTC"`
	AttNorm  int32 `attrib:"attNorm"`
	AttColor int32 `attrib:"attColor"`

	// uniform location for vertesx shader
	UniMatP int32 `uniform:"uniMatP"`
	UniMatV int32 `uniform:"uniMatV"`
	UniMatM int32 `uniform:"uniMatM"`

	// sampler uniform for fragment shader
	UniTex0 int32 `uniform:"uniTex0"`
	UniTex1 int32 `uniform:"uniTex1"`

	UniColors  int32 `uniform:"uniColors"`
	UniClip2D  int32 `uniform:"uniClip2D"`
	UniTexSize int32 `uniform:"uniTexSize"`
}

func newProgram(id uint32) *Program {
	gl.UseProgram(id)
	DbgCheckError()
	p := new(Program)
	p.ID = id
	sv := reflect.ValueOf(p).Elem()
	for i := 0; i < sv.NumField(); i++ {
		ft := sv.Type().Field(i)
		if ft.Type.Kind() != reflect.Int32 {
			continue
		}
		tag := ft.Tag.Get("uniform")
		if tag != "" {
			loc := gl.GetUniformLocation(id, gl.Str(tag+"\x00"))
			sv.Field(i).SetInt(int64(loc))
			// auto assign sampler id
			if strings.HasPrefix(tag, "tex") {
				if x, err := strconv.ParseInt(strings.TrimPrefix(tag, "tex"), 10, 32); err == nil {
					gl.Uniform1i(loc, int32(x))
				}
			}
			continue
		}
		tag = ft.Tag.Get("attrib")
		if tag != "" {
			sv.Field(i).SetInt(int64(gl.GetAttribLocation(id, gl.Str(tag+"\x00"))))
			continue
		}
		sv.Field(i).SetInt(-1)
	}
	return p
}

// UseProgram use the program, same as gl.UseProgram(p.ID)
func (p *Program) UseProgram() {
	//dbg.Logln(p.ID)
	//DbgCheckError()
	gl.UseProgram(p.ID)
	DbgCheckError()
}

// IsCurrentProgram reports whether the program is in use
func (p *Program) IsCurrentProgram() bool {
	var x int32
	gl.GetIntegerv(gl.CURRENT_PROGRAM, &x)
	DbgCheckError()
	return p.ID != 0 && int32(p.ID) == x
}

// SetMVPfv set model, view and projection matrix. i.e. SetMVP(matModel,nil,matProj, false)
func (p *Program) SetMVPfv(model, view, projection *float32, transpose bool) {
	if model != nil {
		if p.UniMatM < 0 {
			panic("p.UniMatM < 0")
		}
		gl.UniformMatrix4fv(p.UniMatM, 1, transpose, model)
		DbgCheckError()
	}
	if view != nil {
		if p.UniMatV < 0 {
			panic("p.UniMatView < 0")
		}
		gl.UniformMatrix4fv(p.UniMatV, 1, transpose, view)
		DbgCheckError()
	}
	if projection != nil {
		if p.UniMatP < 0 {
			panic("p.UniMatProj < 0")
		}
		gl.UniformMatrix4fv(p.UniMatP, 1, transpose, projection)
		DbgCheckError()
	}
}

// SetMVP set model, view and projection matrix.
func (p *Program) SetMVP(model, view, projection Mat4, transpose bool) {
	p.SetMVPfv(&model[0], &view[0], &projection[0], transpose)
}

// LoadMVPStack load model, view and projection matrix from stack
func (p *Program) LoadMVPStack() {
	p.SetMVPfv(StackMatM.GetFloatPtr(), StackMatV.GetFloatPtr(), StackMatP.GetFloatPtr(), false)
}

// LoadVPStack load view and projection matrix from stack
func (p *Program) LoadVPStack() {
	p.SetMVPfv(nil, StackMatV.GetFloatPtr(), StackMatP.GetFloatPtr(), false)
}

// LoadMStack load model matrix from stack
func (p *Program) LoadMStack() {
	p.SetMVPfv(StackMatM.GetFloatPtr(), nil, nil, false)
}

// LoadClip2DStack load clip rect from StackClip2D
func (p *Program) LoadClip2DStack() {
	rect := StackClip2D.Peek()
	//dbg.Logf("rect0=%v\n", rect)
	// convet to NDC
	matp := StackMatP.Get()
	lt := Rect{rect[0], rect[1], 0, 1}
	lt = matp.MultVec4(lt)
	rb := Rect{rect[2], rect[3], 0, 1}
	rb = matp.MultVec4(rb)
	rect[0], rect[1], rect[2], rect[3] = lt[0], lt[1], rb[0], rb[1]
	//dbg.Logf("rect1=%v\n", rect)
	gl.Uniform4fv(p.UniClip2D, 1, &rect[0])
}

func typeIndex(shaderType uint32) int {
	switch shaderType {
	case gl.VERTEX_SHADER:
		return 0
	case gl.FRAGMENT_SHADER:
		return 1
	default:
		log.Panicf("unsupported shader type %d", shaderType)
	}
	return -1
}

// this function add #version directive and compatible defines for GLSL sources
func preProcShader(s []byte, shaderType uint32) (d []byte) {
	d = []byte(glslVersion)

	if shaderType == gl.VERTEX_SHADER {
		d = append(d, []byte(glslIncVert)...)
	} else if shaderType == gl.FRAGMENT_SHADER {
		d = append(d, []byte(glslIncFrag)...)
	} else {
		panic("unkown shader type")
	}
	d = append(d, s...)
	return d
}

func loadShader(name string) (uint32, error) {
	if x, ok := scache[name]; ok {
		return x, nil
	}
	shaderType, ok := exttype[filepath.Ext(name)]
	if !ok {
		shaderType = gl.VERTEX_SHADER
	}
	buf, err := store.ReadFile("shader/" + name)
	if err != nil {
		return 0, err
	}
	buf = preProcShader(buf, shaderType)
	buf = append(buf, '\x00')
	x, err := compileShader(name, string(buf), shaderType)
	if err == nil {
		scache[name] = x
	}
	return x, err
}

func compileShader(name, source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		log = regexp.MustCompile(`(\:)\s*0(\:\s*[0-9]+\:)`).ReplaceAllString(log, `${1} `+name+`${2}`)
		return 0, fmt.Errorf("failed to compile:\n-----\n%v\n-----\n%v", source, log)
	}
	return shader, nil
}

func link(shaders []uint32) (program uint32, err error) {
	program = gl.CreateProgram()
	DbgCheckError()
	defer func() {
		if err != nil {
			gl.DeleteProgram(program)
			program = 0
		}
	}()

	for _, shader := range shaders {
		gl.AttachShader(program, shader)
		DbgCheckError()
	}
	gl.LinkProgram(program)
	DbgCheckError()

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var len int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &len)

		log := strings.Repeat("\x00", int(len+1))
		gl.GetProgramInfoLog(program, len, nil, gl.Str(log))

		err = fmt.Errorf("failed to link program: %v", log)
	}
	return
}

// LoadProgram load and link shaders into program, cached in memory.
func LoadProgram(files ...string) (p *Program, err error) {
	if len(files) == 0 {
		return nil, errors.New("no source files")
	}
	sort.Strings(files)
	pname := strings.Join(files, "|")
	var ok bool
	if p, ok = pcache[pname]; ok {
		return p, nil
	}
	var shaders []uint32
	for _, file := range files {
		var x uint32
		x, err = loadShader(file)
		if err != nil {
			return nil, err
		}
		shaders = append(shaders, x)
	}

	id, err := link(shaders)
	if err == nil {
		p = newProgram(id)
		pcache[pname] = p
	}
	return
}

// MustLoadProgram load and link shaders into program, cached in memory.
// panics if failed.
func MustLoadProgram(files ...string) *Program {
	x, err := LoadProgram(files...)
	if err == nil {
		return x
	}
	log.Printf("GL_SHADING_LANGUAGE_VERSION=%s\n", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	log.Panicf("LoadProgram %v:\n%v", files, err)
	return nil
}
