package jex

import (
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"
)

var errJexFormat = errors.New("Not Jex format")
var errJexVersion = errors.New("Unsupported Jex version")

func golangTime(t float64) time.Time {
	sec, frac := math.Modf(t)
	nano := int64(frac * 1000000000)
	//	console.Debug(t, sec, nano)
	return time.Unix(int64(sec), nano)
}

func jexTime(t time.Time) float64 {
	sec, nano := t.Unix(), t.UnixNano()%1e9
	return float64(sec) + float64(nano)*0.000000001
}

const (
	cNil     byte = 0x00
	cTrue    byte = 0x01
	cFalse   byte = 0x02
	cInt8    byte = 0x03
	cUInt8   byte = 0x04
	cInt64   byte = 0x05
	cUInt64  byte = 0x06
	cFloat32 byte = 0x07
	cFloat64 byte = 0x08
	cString  byte = 0x09
	cBytes   byte = 0x0A // 其它语言(如swift)里string和字节数组可能不同, 所以要区分
	cObject  byte = 0x0B
	cVec2    byte = 0x0C
	cVec3    byte = 0x0D
	cVec4    byte = 0x0E
	cMat3    byte = 0x0F
	cMat4    byte = 0x10
	cTime    byte = 0x11

	cHasChildren byte = 0x80
	cHasName     byte = 0x40
	cTypeMask    byte = 0x3F
)

// 供对象实现的接口, 用自己的方法写入Jex数据块
// Jex数据块要求
type Marshaler interface {
	MarshalJex(w io.Writer) error
}

// 供对象实现的接口, 用自己的方法读取Jex数据块
type Unmarshaler interface {
	UnmarshalJex(r io.Reader) error
}

var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()
var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

// 辅助函数, 写入一个字节
func WriteByte(w io.Writer, c byte) error {
	_, err := w.Write([]byte{c})
	return err
}

// 辅助函数, 读取一个字节
func ReadByte(r io.Reader) (byte, error) {
	var buf [1]byte
	_, err := r.Read(buf[0:1])
	return buf[0], err
}

// 辅助函数, 写入一个变长整数
func WriteUvarint(w io.Writer, x uint64) error {
	buf := [binary.MaxVarintLen64]byte{}
	n := binary.PutUvarint(buf[:], x)
	_, err := w.Write(buf[0:n])
	return err
}

type byteReader struct {
	r io.Reader
}

func (br byteReader) ReadByte() (byte, error) {
	var buf [1]byte
	_, err := br.r.Read(buf[0:1])
	return buf[0], err
}

// 辅助函数, 读取一个变长整数
func ReadUvarint(r io.Reader) (uint64, error) {
	return binary.ReadUvarint(byteReader{r})
}

// 辅助函数, 写入一个变长整数
func WriteVarint(w io.Writer, x int64) error {
	buf := [binary.MaxVarintLen64]byte{}
	n := binary.PutVarint(buf[:], x)
	_, err := w.Write(buf[0:n])
	return err
}

// 辅助函数, 读取一个变长整数
func ReadVarint(r io.Reader) (int64, error) {
	return binary.ReadVarint(byteReader{r})
}

// 辅助函数, 写入一段二进制数据
func WriteBytes(w io.Writer, s []byte) error {
	if err := WriteUvarint(w, uint64(len(s))); err != nil {
		return err
	}
	_, err := w.Write(s)
	return err
}

func ReadBytes(r io.Reader) ([]byte, error) {
	n, err := ReadUvarint(r)
	if err != nil || n == 0 {
		return nil, err
	}
	buf := make([]byte, n)
	_, err = r.Read(buf)
	return buf, err
}

// 辅助函数, 写入一个字符串
func WriteString(w io.Writer, s string) error {
	return WriteBytes(w, []byte(s))
}

// 辅助函数, 读取一个字符串
func ReadString(r io.Reader) (string, error) {
	buf, err := ReadBytes(r)
	return string(buf), err
}

// 写裸数据. name 可以为空
func MarshalNaked(w io.Writer, in interface{}, name string) error {
	var nameFlag byte
	if name != "" {
		nameFlag = cHasName
	}
	data := reflect.ValueOf(in)
	for data.Kind() == reflect.Ptr && !data.IsNil() && !data.Type().Implements(marshalerType) {
		data = reflect.Indirect(data)
	}
	in = data.Interface()
	switch v := in.(type) {
	case nil:
		if err := WriteByte(w, nameFlag|cNil); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return nil
	case Marshaler:
		// 实际上就是一个字节数组
		if err := WriteByte(w, nameFlag|cBytes); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return v.MarshalJex(w)
	case bool:
		f := nameFlag
		if v {
			f |= cTrue
		} else {
			f |= cFalse
		}
		if err := WriteByte(w, f); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return nil
	case int:
		if err := WriteByte(w, nameFlag|cInt64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteVarint(w, int64(v))
	case int64:
		if err := WriteByte(w, nameFlag|cInt64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteVarint(w, int64(v))
	case int32:
		if err := WriteByte(w, nameFlag|cInt64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteVarint(w, int64(v))
	case uint:
		if err := WriteByte(w, nameFlag|cUInt64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteUvarint(w, uint64(v))
	case uint64:
		if err := WriteByte(w, nameFlag|cUInt64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteUvarint(w, uint64(v))
	case uint32:
		if err := WriteByte(w, nameFlag|cUInt64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteUvarint(w, uint64(v))
	case string:
		if err := WriteByte(w, nameFlag|cString); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteString(w, v)
	case []byte:
		// console.Debug("[]byte")
		if err := WriteByte(w, nameFlag|cBytes); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteBytes(w, v)
	case int8:
		if err := WriteByte(w, nameFlag|cInt8); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteByte(w, uint8(v))
	case uint8:
		if err := WriteByte(w, nameFlag|cUInt8); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return WriteByte(w, v)
	case float32:
		if err := WriteByte(w, nameFlag|cFloat32); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return binary.Write(w, binary.LittleEndian, v)
	case float64:
		if err := WriteByte(w, nameFlag|cFloat64); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return binary.Write(w, binary.LittleEndian, v)
	case time.Time:
		if err := WriteByte(w, nameFlag|cTime); err != nil {
			return err
		}
		if name != "" {
			if err := WriteString(w, name); err != nil {
				return err
			}
		}
		return binary.Write(w, binary.LittleEndian, jexTime(v))
	default:
		switch data.Kind() {
		case reflect.Map:
			keys := data.MapKeys()
			n := len(keys)
			//			if n == 0 {
			//				err := WriteByte(w, nameFlag|cNil)
			//				if err != nil {
			//					return err
			//				}
			//				if name != "" {
			//					if err := WriteString(w, name); err != nil {
			//						return err
			//					}
			//				}
			//				return nil
			//			}
			if err := WriteByte(w, nameFlag|cHasChildren); err != nil {
				return err
			}
			if name != "" {
				if err := WriteString(w, name); err != nil {
					return err
				}
			}
			if err := WriteUvarint(w, uint64(n)); err != nil {
				return err
			}
			for _, k := range keys {
				if err := MarshalNaked(w, data.MapIndex(k).Interface(), k.String()); err != nil {
					return err
				}
			}
		case reflect.Slice, reflect.Array:
			n := data.Len()
			//			if n == 0 {
			//				err := WriteByte(w, nameFlag|cNil)
			//				if err != nil {
			//					return err
			//				}
			//				if name != "" {
			//					if err := WriteString(w, name); err != nil {
			//						return err
			//					}
			//				}
			//				return nil
			//			}
			if err := WriteByte(w, nameFlag|cHasChildren); err != nil {
				return err
			}
			if name != "" {
				if err := WriteString(w, name); err != nil {
					return err
				}
			}
			if err := WriteUvarint(w, uint64(n)); err != nil {
				return err
			}
			for i := 0; i < n; i++ {
				if err := MarshalNaked(w, data.Index(i).Interface(), ""); err != nil {
					return err
				}
			}
		case reflect.Struct:
			var st = data.Type()
			stinfo, err := getStructInfo(st)
			if err != nil {
				return err
			}
			n := len(stinfo.FieldsList)
			//			if n == 0 {
			//				err := WriteByte(w, nameFlag|cNil)
			//				if err != nil {
			//					return err
			//				}
			//				if name != "" {
			//					if err := WriteString(w, name); err != nil {
			//						return err
			//					}
			//				}
			//				return nil
			//			}

			if err := WriteByte(w, nameFlag|cHasChildren); err != nil {
				return err
			}
			if name != "" {
				if err = WriteString(w, name); err != nil {
					return err
				}
			}
			if err := WriteUvarint(w, uint64(n)); err != nil {
				return err
			}
			// Jex特点是字段有序(虽然非强制), 所以按顺序来
			for _, fi := range stinfo.FieldsList {
				if err := MarshalNaked(w, data.Field(fi.Num).Interface(), fi.Key); err != nil {
					return err
				}
			}
		default:
			return errors.New(fmt.Sprintf("Unsupported data type: %s", reflect.ValueOf(in).Type().String()))
		}
		return nil
	}
	// return nil
}

// 写文件头
func WriteHeader(w io.Writer, zip bool) error {
	// 目前的版本是 0
	if err := WriteByte(w, 0xC0); err != nil {
		return err
	}
	// 非压缩格式
	var b1 byte
	if zip {
		b1 = 0x6a
	} else {
		b1 = 0x4a
	}
	if err := WriteByte(w, b1); err != nil {
		return err
	}
	return nil
}

// 写标准Jex文件
// 标准的Jex格式需要有文件头, 以识别是否压缩.
// 此外, 标准的Jex只有一个根节点, 其它内容需要放到根节点下.
func Marshal(w io.Writer, in interface{}, zip bool) error {
	if err := WriteHeader(w, zip); err != nil {
		return err
	}
	if !zip {
		return MarshalNaked(w, in, "")
	}
	zw, err := zlib.NewWriterLevel(w, zlib.DefaultCompression)
	if err != nil {
		return err
	}
	defer zw.Close()
	return MarshalNaked(zw, in, "")
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "jex: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "jex: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "jex: Unmarshal(nil " + e.Type.String() + ")"
}

func _readFlagName(r io.Reader) (flag byte, name string, err error) {
	flag, err = ReadByte(r)
	if err != nil {
		return
	}
	if flag&cHasName != 0 {
		name, err = ReadString(r)
	}
	return
}

func _discard(r io.Reader, n uint) error {
	if n == 0 {
		return nil
	}

	var tmp = make([]byte, n)
	_, err := r.Read(tmp)
	return err
}

func _discardVariant(r io.Reader) (err error) {
	_, err = ReadUvarint(r)
	return
}

func _discardString(r io.Reader) (err error) {
	n, err := ReadUvarint(r)
	if err != nil {
		return err
	}
	return _discard(r, uint(n))
}

func _unmarshal_discard(r io.Reader, flag byte) error {
	switch flag & cTypeMask {
	case cNil, cTrue, cFalse:
		break
	case cInt8, cUInt8:
		if err := _discard(r, 1); err != nil {
			return err
		}
	case cInt64, cUInt64, cObject:
		if err := _discardVariant(r); err != nil {
			return err
		}
	case cFloat32:
		if err := _discard(r, 4); err != nil {
			return err
		}
	case cFloat64:
		if err := _discard(r, 8); err != nil {
			return err
		}
	case cString, cBytes:
		if err := _discardString(r); err != nil {
			return err
		}
	case cVec2:
		if err := _discard(r, 8); err != nil {
			return err
		}
	case cVec3:
		if err := _discard(r, 12); err != nil {
			return err
		}
	case cVec4: //byte = 0x0E
		if err := _discard(r, 16); err != nil {
			return err
		}
	case cMat3:
		if err := _discard(r, 36); err != nil {
			return err
		}
	case cMat4:
		if err := _discard(r, 64); err != nil {
			return err
		}
	case cTime:
		if err := _discard(r, 8); err != nil {
			return err
		}
	default:
		return errors.New(fmt.Sprintf("Unsupported type: %d", flag))
	}

	if flag&cHasChildren != 0 {
		n, err := ReadUvarint(r)
		if err != nil {
			return err
		}

		for i := 0; i < int(n); i++ {
			flag1, _, err1 := _readFlagName(r)
			err1 = _unmarshal_discard(r, flag1)
			if err1 != nil {
				return err1
			}
		}
	}
	return nil

}

func _unmarshal_field(r io.Reader, outPtr reflect.Value, flag byte) (ret error) {
	if outPtr.Interface() == nil {
		return _unmarshal_discard(r, flag)
	}
	//	defer func() {
	//		if e := recover(); e != nil {
	//			ret = errors.New(fmt.Sprint(e))
	//		}
	//	}()

	//outPtr := reflect.ValueOf(out)
	if outPtr.Kind() != reflect.Ptr || outPtr.IsNil() {
		return &InvalidUnmarshalError{outPtr.Type()}
	}
	outValue := outPtr.Elem()
	outType := outPtr.Type().Elem()

	switch flag & cTypeMask {
	case cNil:
		outValue.Set(reflect.Zero(outType))
	case cTrue:
		outValue.SetBool(true)
	case cFalse:
		outValue.SetBool(false)
	case cInt8:
		tmp, err := ReadByte(r)
		if err != nil {
			return err
		}
		outValue.SetInt(int64(int8(tmp)))
	case cUInt8:
		tmp, err := ReadByte(r)
		if err != nil {
			return err
		}
		outValue.SetUint(uint64(tmp))
	case cInt64:
		tmp, err := ReadVarint(r)
		if err != nil {
			return err
		}
		outValue.SetInt(tmp)
	case cUInt64, cObject:
		tmp, err := ReadUvarint(r)
		if err != nil {
			return err
		}
		outValue.SetUint(tmp)
	case cFloat32:
		var tmp float32
		if err := binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return err
		}
		outValue.SetFloat(float64(tmp))
	case cFloat64:
		var tmp float64
		if err := binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return err
		}
		outValue.SetFloat(tmp)
	case cString:
		tmp, err := ReadString(r)
		if err != nil {
			return err
		}
		outValue.SetString(tmp)
	case cBytes:
		tmp, err := ReadBytes(r)
		if err != nil {
			return err
		}
		outValue.SetBytes(tmp)
	case cVec2: ///byte = 0x0C
		return errors.New("vec2 not supported yet")
	case cVec3: // byte = 0x0D
		return errors.New("vec3 not supported yet")
	case cVec4: //byte = 0x0E
		return errors.New("vec4 not supported yet")
	case cMat3: //byte = 0x0F
		return errors.New("mat3 not supported yet")
	case cMat4: // byte = 0x10
		return errors.New("mat4 not supported yet")
	case cTime:
		//	console.Debug("cTime")
		var tmp float64
		if err := binary.Read(r, binary.LittleEndian, &tmp); err != nil {
			return err
		}
		t := golangTime(tmp)
		outValue.Set(reflect.ValueOf(t))
	//	console.Debug(outValue, reflect.ValueOf(t))
	default:
		return errors.New(fmt.Sprintf("Unsupported type: %d", flag))
	}

	if flag&cHasChildren != 0 {
		n64, err := ReadUvarint(r)
		n := int(n64)
		if err != nil {
			ret = err
			return
		}
		switch outType.Kind() {
		case reflect.Array, reflect.Slice:
			if n == 0 {
				outValue.Set(reflect.Zero(outType))
				return nil
			}
			if outType.Kind() == reflect.Slice {
				outValue.Set(reflect.MakeSlice(outType, n, n))
			}
			for i := 0; i < int(n); i++ {
				flag1, _, err1 := _readFlagName(r)
				if err1 != nil {
					return err1
				}
				if i < outValue.Len() {
					err1 = _unmarshal_field(r, outValue.Index(i).Addr(), flag1)
					if err1 != nil {
						return err1
					}
				}
			}
			for i := int(n); i < outValue.Len(); i++ {
				outValue.Index(i).Set(reflect.Zero(outType.Elem()))
			}
		case reflect.Map:
			if n == 0 {
				outValue.Set(reflect.Zero(outType))
				return nil
			}
			outValue.Set(reflect.MakeMap(outType))
			for i := 0; i < int(n); i++ {
				flag1, key, err1 := _readFlagName(r)
				if err1 != nil {
					return err1
				}
				tmpPtr := reflect.New(outType.Elem())
				err1 = _unmarshal_field(r, tmpPtr, flag1)
				if err1 != nil {
					return err1
				}
				outValue.SetMapIndex(reflect.ValueOf(key), tmpPtr.Elem())
			}
		case reflect.Struct:
			stinfo, err := getStructInfo(outType)
			if err != nil {
				return err
			}
			for i := 0; i < int(n); i++ {
				flag1, name1, err1 := _readFlagName(r)
				if err1 != nil {
					return err1
				}
				fi, ok := stinfo.FieldsMap[name1]
				if !ok {
					if err := _unmarshal_discard(r, flag1); err != nil {
						return err
					}
				}
				f := outValue.Field(fi.Num)
				for f.Kind() == reflect.Ptr && f.IsNil() {
					f.Set(reflect.Zero(f.Type().Elem()))
					f = f.Elem()
				}
				if f.Kind() != reflect.Ptr {
					f = f.Addr()
				}
				err1 = _unmarshal_field(r, f, flag1)
				if err1 != nil {
					return err1
				}
			}
		default:
			return errors.New(fmt.Sprintf("Not a container type: %s", outType.String()))
		}
	} else if flag&cTypeMask != cBytes {
		switch outType.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map:
			outValue.Set(reflect.Zero(outType))
			return nil
		}
	}
	return
}

// 读裸数据
func UnmarshalNaked(r io.Reader, out interface{}) (name string, err error) {
	flag, name, err := _readFlagName(r)
	if err != nil {
		return
	}
	err = _unmarshal_field(r, reflect.ValueOf(out), flag)
	return
}

// 读文件头
func ReadHeader(r io.Reader) (ver int, zip bool, err error) {
	b0, err := ReadByte(r)
	if err != nil {
		return
	}
	if (b0 & 0xF0) != 0xC0 {
		err = errJexFormat
		return
	}

	b1, err := ReadByte(r)
	if err != nil {
		return
	}
	if b1 != 0x4a && b1 != 0x6a {
		err = errJexFormat
		return
	}
	ver = int(b0 & 0x0F)
	zip = b1 == 0x6a
	return
}

func Unmarshal(r io.Reader, out interface{}) error {
	ver, zip, err := ReadHeader(r)
	if err != nil {
		return err
	}
	if ver != 0 {
		return errJexVersion
	}
	if !zip {
		_, err = UnmarshalNaked(r, out)
		return err
	}
	zr, err := zlib.NewReader(r)
	if err != nil {
		return err
	}
	_, err = UnmarshalNaked(zr, out)
	return err
}

// --------------------------------------------------------------------------
// Maintain a mapping of keys to structure field indexes

type structInfo struct {
	FieldsMap  map[string]fieldInfo
	FieldsList []fieldInfo
	Zero       reflect.Value
}

type fieldInfo struct {
	Key       string
	Num       int
	OmitEmpty bool
}

var structMap = make(map[reflect.Type]*structInfo)
var structMapMutex sync.RWMutex

func getStructInfo(st reflect.Type) (*structInfo, error) {
	structMapMutex.RLock()
	sinfo, found := structMap[st]
	structMapMutex.RUnlock()
	if found {
		return sinfo, nil
	}
	n := st.NumField()
	fieldsMap := make(map[string]fieldInfo)
	fieldsList := make([]fieldInfo, 0, n)
	for i := 0; i != n; i++ {
		field := st.Field(i)
		if field.PkgPath != "" {
			continue // Private field
		}

		info := fieldInfo{Num: i}

		tag := field.Tag.Get("jex")
		if tag == "" && strings.Index(string(field.Tag), ":") < 0 {
			tag = string(field.Tag)
		}
		if tag == "-" {
			continue
		}

		fields := strings.Split(tag, ",")
		if len(fields) > 1 {
			for _, flag := range fields[1:] {
				switch flag {
				//case "omitempty":
				//	info.OmitEmpty = true
				default:
					msg := fmt.Sprintf("Unsupported flag %q in tag %q of type %s", flag, tag, st)
					panic(msg)
				}
			}
			tag = fields[0]
		}

		if tag != "" {
			if tag != "*" {
				info.Key = tag
			}
		} else {
			info.Key = strings.ToLower(field.Name)
		}

		if _, found = fieldsMap[info.Key]; found {
			msg := "Duplicated key '" + info.Key + "' in struct " + st.String()
			return nil, errors.New(msg)
		}

		fieldsList = append(fieldsList, info)
		fieldsMap[info.Key] = info
	}
	sinfo = &structInfo{
		fieldsMap,
		fieldsList,
		reflect.New(st).Elem(),
	}
	structMapMutex.Lock()
	structMap[st] = sinfo
	structMapMutex.Unlock()
	return sinfo, nil
}
