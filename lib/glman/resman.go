package glman

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"tetra/internal/gl"
	"tetra/lib/dbg"
)

// OpenGL object types
const (
	tTexture = iota
	tVertexArray
	tBuffer
	maxType
)

// Provider is callback to load resources
type Provider func(name string) *Res

// Providers
var (
	ProvideTexture Provider
	//ProvideVertexArray Provider
	//ProvideBuffer      Provider
)

var (
	pendingDestroys [maxType][]*sharedRes
	mutexResMan     = make(chan int, 1)

	resCaches [maxType]map[string]*sharedRes
)

func init() {
	for i := range resCaches {
		resCaches[i] = make(map[string]*sharedRes)
	}
}

// Routine will destory all pending object
func Routine() {
	mutexResMan <- 1
	for typ, s := range pendingDestroys {
		if len(s) == 0 {
			continue
		}
		for i, r := range s {
			if r.name != "" {
				delete(resCaches[typ], r.name)
			}
			switch typ {
			case tTexture:
				dbg.Logf("glgc: gl.DeleteTextures %d\n", r.id)
				gl.DeleteTextures(1, &r.id)
				DbgCheckError()
			case tVertexArray:
				dbg.Logf("glgc: gl.DeleteVertexArrays %d\n", r.id)
				gl.DeleteVertexArrays(1, &r.id)
				DbgCheckError()
			case tBuffer:
				dbg.Logf("glgc: gl.DeleteBuffers %d\n", r.id)
				gl.DeleteBuffers(1, &r.id)
				DbgCheckError()
			default:
				panic(fmt.Sprintf("destroy type %d", i))
			}
			s[i] = nil
		}
		pendingDestroys[typ] = s[:0]
	}
	<-mutexResMan
}

// sharedRes is wrapper of OpenGL resource id, with auto management feature.
type sharedRes struct {
	id   uint32
	typ  int
	ref  int32
	ac   uint32
	name string
}

// Res is pointer to resource
type Res struct {
	s *sharedRes
}

func release(r *Res) {
	s := r.s
	if s == nil {
		return
	}
	x := atomic.AddInt32(&s.ref, -1)
	dbg.Logf("glgc: release %v\n", r)
	if x == 0 {
		// no reference any more, free it
		if s.typ >= maxType {
			panic(fmt.Sprintf("free: %s", r))
		}
		mutexResMan <- 1
		pendingDestroys[s.typ] = append(pendingDestroys[s.typ], s)
		<-mutexResMan
		r.s = nil
	}
}

// Release reference explicitly
func (r *Res) Release() {
	release(r)
}

// ID of OpenGL resource
func (r *Res) ID() uint32 {
	if r.s == nil {
		return 0
	}
	return r.s.id
}

// Type of the resource, for debugging
func (r *Res) Type() string {
	if r.s == nil {
		return "Invalid"
	}
	switch r.s.typ {
	case tTexture:
		return "Texture"
	case tBuffer:
		return "Buffer"
	case tVertexArray:
		return "VetexArray"
	default:
		return "Unkown"
	}
}

// NumRef reports reference count
func (r *Res) NumRef() int {
	if r.s == nil {
		return 0
	}
	return int(atomic.LoadInt32(&r.s.ref))
}

// Name of the resource
func (r *Res) Name() string {
	if r.s == nil {
		return ""
	}
	return r.s.name
}

func (r *Res) String() string {
	return fmt.Sprintf("[gl %s id=%d ref=%d \"%s\"]", r.Type(), r.ID(), r.NumRef(), r.Name())
}

func ref(s *sharedRes) *Res {
	if s == nil || s.id == 0 {
		return nil
	}
	atomic.AddInt32(&s.ref, 1)
	r := &Res{s}
	runtime.SetFinalizer(r, release)
	return r
}

// GenTexture is wrapper for gl.GenTextures
func GenTexture(name string) *Res {
	s := &sharedRes{typ: tTexture, name: name}
	gl.GenTextures(1, &s.id)
	DbgCheckError()
	return ref(s)
}

// LoadTexture load and cache texture by name, must set ProvideTexture before call this function.
func LoadTexture(name string) *Res {
	if name == "" {
		return nil
	}
	if s, ok := resCaches[tTexture][name]; ok {
		return ref(s)
	}
	if ProvideTexture == nil {
		return nil
	}
	r := ProvideTexture(name)
	if r == nil {
		return nil
	}
	resCaches[tTexture][name] = r.s
	return r
}

// GenBuffer is wrapper for gl.GenBuffers
func GenBuffer(name string) *Res {
	s := &sharedRes{typ: tBuffer, name: name}
	gl.GenBuffers(1, &s.id)
	DbgCheckError()
	return ref(s)
}

// GenVertexArray is wrapper for gl.GenVertexArrays
func GenVertexArray(name string) *Res {
	s := &sharedRes{typ: tVertexArray, name: name}
	gl.GenVertexArrays(1, &s.id)
	DbgCheckError()
	return ref(s)
}
