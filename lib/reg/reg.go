// Package reg provide a key-value data storage.
//
// the storage is actually a JSON tree
//
// keys are strings, can be separate into multiple level, like the path of filesystem.
// the separator is '/', leading and ending '/' is ignored, continious '/' is treat as
// single '/', so "/foo//bar/" is same as "foo/bar". keys are case sensitive.
//
// internal values should be valid JSON, i.e. one of nil, bool, float64, string, []interface{}
// or map[string]interface{}. but we also provide interfaces for compatible types, even a
// universal Marshal method. when putting non-standard values into storage, they are converted to
// JSON types.
package reg

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

var (
	errNotFound    = errors.New("Not found")
	errWrongType   = errors.New("Wrong type")
	errEmpyKey     = errors.New("Key is empty")
	errInternal    = errors.New("Internal error")
	errEndOfObject = errors.New("End of object: '}'")
	errEndOfArray  = errors.New("End of array: ']'")
)

func decodeJSON(dec *json.Decoder) (j interface{}, err error) {
	t, err := dec.Token()
	if err != nil {
		return
	}
	switch x := t.(type) {
	case nil:
		break
	case bool:
		j = x
	case float64:
		j = x
	case json.Number:
		j, err = x.Float64()
	case string:
		j = x
	case json.Delim:
		switch x {
		case '{':
			obj := make(map[string]interface{})
			for {
				var keyTk json.Token
				keyTk, err = dec.Token()
				if err != nil {
					break
				}
				keyDl, ok := keyTk.(json.Delim)
				if ok {
					if keyDl == '}' {
						j = obj
						break
					}
				}
				key, ok := keyTk.(string)
				if !ok {
					panic(errInternal) // should handled by decoder
				}
				var val interface{}
				val, err = decodeJSON(dec)
				if err != nil {
					break
				}
				obj[key] = val
			}
		case '}':
			err = errEndOfObject
		case '[':
			var arr []interface{}
			for {
				var val interface{}
				val, err = decodeJSON(dec)
				if err != nil {
					if err == errEndOfArray {
						err = nil
						j = arr
					}
					break
				}
				arr = append(arr, val)
			}
		case ']':
			err = errEndOfArray
		default:
			err = errInternal
		}
	default:
		err = errInternal
	}
	return
}

// Reg is key-value database.
type Reg struct {
	j interface{}
	m sync.RWMutex
}

// DecodeReader decode JSON tree from reader
func DecodeReader(r io.Reader) (*Reg, error) {
	j, err := decodeJSON(json.NewDecoder(r))
	if err != nil {
		return nil, err
	}
	rg := new(Reg)
	rg.j = j
	return rg, nil
}

// Decode JSON tree
func Decode(data []byte) (*Reg, error) {
	return DecodeReader(bytes.NewReader(data))
}

// Encode to JSON tree
func (rg *Reg) Encode() ([]byte, error) {
	rg.m.RLock()
	defer rg.m.RUnlock()
	return json.Marshal(rg.j)
}

// EncodeIndent like Encode, with indentation
func (rg *Reg) EncodeIndent(prefix, indent string) ([]byte, error) {
	rg.m.RLock()
	defer rg.m.RUnlock()
	return json.MarshalIndent(rg.j, prefix, indent)
}

func (rg *Reg) get(path string) (interface{}, error) {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil, errEmpyKey
	}

	j := rg.j
	if j == nil {
		return nil, errNotFound
	}
	var key string
	for path != "" {
		pos := strings.IndexByte(path, '/')
		if pos == 0 {
			// "/foo//bar"
			path = path[1:]
			continue
		}
		if pos == -1 {
			key = path
			path = ""
		} else {
			key = path[:pos]
			path = path[pos+1:]
		}
		// fmt.Println(key)
		if j == nil {
			return nil, errNotFound
		}
		if obj, ok := j.(map[string]interface{}); ok {
			if k, ok := obj[key]; ok {
				j = k
				continue
			}
			return nil, errNotFound

		}
		return nil, errWrongType
	}
	return j, nil
}

func (rg *Reg) set(path string, x interface{}) error {
	backPath := path
	path = strings.Trim(path, "/")
	if path == "" {
		// for safety we don't allow set to root element
		return errEmpyKey
	}
	rg.m.Lock()
	defer rg.m.Unlock()

	var obj map[string]interface{}
	var ok bool
	if obj, ok = rg.j.(map[string]interface{}); !ok {
		if rg.j != nil {
			log.Printf(`set "%s": convert root to object cause data lost`, backPath)
		}
		obj = make(map[string]interface{})
		rg.j = obj
	}
	var key string
	for {
		pos := strings.IndexByte(path, '/')
		if pos == 0 {
			// "/foo//bar"
			path = path[1:]
			continue
		}
		if pos == -1 {
			key = path
			path = ""
		} else {
			key = path[:pos]
			path = path[pos+1:]
		}

		if path == "" {
			obj[key] = x
			// fmt.Printf("%#v\n", rg.j)
			return nil
		}
		parent := obj
		var tmp interface{}
		if tmp, ok = parent[key]; !ok {
			obj = make(map[string]interface{})
			parent[key] = obj
		} else if obj, ok = tmp.(map[string]interface{}); !ok {
			if rg.j != nil {
				log.Printf(`set "%s": convert "%s" to object cause data lost`, backPath, key)
			}
			obj = make(map[string]interface{})
			parent[key] = obj
		}
	}
}

// GetUnmarshal use josn.Unmarshal to get data at key
func (rg *Reg) GetUnmarshal(key string, v interface{}) error {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return err
	}

	b, err := json.Marshal(j)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

// SetMarshal use josn.Marshal to set data to key
func (rg *Reg) SetMarshal(key string, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	j, err := decodeJSON(json.NewDecoder(bytes.NewReader(b)))
	if err != nil {
		return err
	}

	rg.set(key, j)
	return nil
}

// GetBool get boolean value
func (rg *Reg) GetBool(key string) (x bool, err error) {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return
	}
	// we don't convert other types, because not well defined
	if v, ok := j.(bool); ok {
		x = v
	} else {
		err = errWrongType
	}
	return
}

// SetBool set boolean value
func (rg *Reg) SetBool(key string, x bool) error {
	return rg.set(key, x)
}

// GetString get string value
func (rg *Reg) GetString(key string) (x string, err error) {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return
	}
	switch v := j.(type) {
	case nil:
		x = "null"
	case bool:
		if v {
			x = "true"
		} else {
			x = "false"
		}
	//case float64:
	//x = strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		x = v
	default:
		var b []byte
		b, err = json.Marshal(v)
		x = string(b)
	}
	return
}

// SetString set string value
func (rg *Reg) SetString(key string, x string) error {
	return rg.set(key, x)
}

// GetBytes get []byte value decode from base64
func (rg *Reg) GetBytes(key string) (x []byte, err error) {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return
	}
	var s string
	var ok bool
	if s, ok = j.(string); !ok {
		return nil, errWrongType
	}

	x, err = base64.RawURLEncoding.DecodeString(s)
	return
}

// SetBytes set []byte value encode into base64
func (rg *Reg) SetBytes(key string, x []byte) error {
	return rg.set(key, base64.RawURLEncoding.EncodeToString(x))
}

func (rg *Reg) getFloat(key string, bitSize int) (x float64, err error) {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return
	}
	switch v := j.(type) {
	case nil:
		break
	case bool:
		if v {
			x = 1
		}
	case float64:
		x = v
	case string:
		x, err = strconv.ParseFloat(v, bitSize)
	default:
		err = errInternal
	}
	return
}

func (rg *Reg) getInt(key string, bitSize int) (x int64, err error) {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return
	}
	switch v := j.(type) {
	case nil:
		break
	case bool:
		if v {
			x = 1
		}
	case float64:
		x = int64(v)
	case string:
		x, err = strconv.ParseInt(v, 10, bitSize)
	default:
		err = errInternal
	}
	return
}

func (rg *Reg) getUint(key string, bitSize int) (x uint64, err error) {
	rg.m.RLock()
	defer rg.m.RUnlock()

	j, err := rg.get(key)
	if err != nil {
		return
	}
	switch v := j.(type) {
	case nil:
		break
	case bool:
		if v {
			x = 1
		}
	case float64:
		x = uint64(v)
	case string:
		x, err = strconv.ParseUint(v, 10, bitSize)
	default:
		err = errInternal
	}
	return
}

// GetFloat64 get float64 value
func (rg *Reg) GetFloat64(key string) (x float64, err error) {
	x, err = rg.getFloat(key, 64)
	return
}

// GetFloat32 get float32 value
func (rg *Reg) GetFloat32(key string) (x float32, err error) {
	var f float64
	f, err = rg.getFloat(key, 32)
	x = float32(f)
	return
}

// GetInt64 get int64 value
func (rg *Reg) GetInt64(key string) (x int64, err error) {
	x, err = rg.getInt(key, 64)
	return
}

// GetInt32 get int32 value
func (rg *Reg) GetInt32(key string) (x int32, err error) {
	var f int64
	f, err = rg.getInt(key, 32)
	x = int32(f)
	return
}

// GetUint64 get int64 value
func (rg *Reg) GetUint64(key string) (x uint64, err error) {
	x, err = rg.getUint(key, 64)
	return
}

// GetUint32 get int32 value
func (rg *Reg) GetUint32(key string) (x uint32, err error) {
	var f uint64
	f, err = rg.getUint(key, 32)
	x = uint32(f)
	return
}

// GetInt get int value
func (rg *Reg) GetInt(key string) (x int, err error) {
	var f int64
	f, err = rg.getInt(key, int(unsafe.Sizeof(x)*8))
	x = int(f)
	return
}

// GetUint get int value
func (rg *Reg) GetUint(key string) (x uint, err error) {
	var f uint64
	f, err = rg.getUint(key, int(unsafe.Sizeof(x)*8))
	x = uint(f)
	return
}

// SetFloat64 set float64 value
func (rg *Reg) SetFloat64(key string, x float64) error {
	return rg.set(key, x)
}

// SetFloat32 set float32 value
func (rg *Reg) SetFloat32(key string, x float32) error {
	return rg.set(key, float64(x))
}

// SetInt32 set int32 value
func (rg *Reg) SetInt32(key string, x int32) error {
	return rg.set(key, float64(x))
}

// SetUint32 set float32 value
func (rg *Reg) SetUint32(key string, x float32) error {
	return rg.set(key, float64(x))
}

// SetInt64 set int64 value
func (rg *Reg) SetInt64(key string, x int64) error {
	if x < math.MinInt32 || x > math.MaxInt32 {
		return rg.set(key, strconv.FormatInt(x, 10))
	}
	return rg.set(key, float64(x))
}

// SetUint64 set uint64 value
func (rg *Reg) SetUint64(key string, x uint64) error {
	if x > math.MaxUint32 {
		return rg.set(key, strconv.FormatUint(x, 10))
	}
	return rg.set(key, float64(x))
}

// SetInt set int value
func (rg *Reg) SetInt(key string, x int) error {
	if x < math.MinInt32 || x > math.MaxInt32 {
		return rg.set(key, strconv.FormatInt(int64(x), 10))
	}
	return rg.set(key, float64(x))
}

// SetUint set uint value
func (rg *Reg) SetUint(key string, x uint) error {
	if x > math.MaxUint32 {
		return rg.set(key, strconv.FormatUint(uint64(x), 10))
	}
	return rg.set(key, float64(x))
}
