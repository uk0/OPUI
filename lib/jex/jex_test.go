package jex

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"testing"
	"time"
)

type TestA struct {
	Int8  int8
	UInt8 uint8
	Int   int
	UInt  uint
	Map   map[string]TestA
	Array [2]float32
	Slice []float64
}

func TestMarshalStruct(t *testing.T) {
	m := map[string]TestA{
		"ASDF": TestA{1, 2, 3, 4, nil, [2]float32{}, nil},
	}
	a := TestA{-123, 234, -456780, 567890, m, [2]float32{3.14, 1.414}, []float64{0.123, -0.000000001}}
	buf := bytes.NewBuffer(nil)
	err := Marshal(buf, &a, false)
	if err != nil {
		t.Fatal(err)
	}
	ioutil.WriteFile("a.jex", buf.Bytes(), 0644)
	var b TestA
	err = Unmarshal(buf, &b)
	if err != nil {
		t.Fatal(err)
	}
	//	t.Logf("%#v", a)
	//	t.Logf("%#v", b)
	if !reflect.DeepEqual(a, b) {
		t.Error("a != b")
	}
}

func TesMarshalSlice(t *testing.T) {
	var slice = []float64{1, 3, 4}
	buf := bytes.NewBuffer(nil)
	err := Marshal(buf, slice, true)
	if err != nil {
		t.Fatal(err)
	}
	//ioutil.WriteFile("b.jex", buf.Bytes(), 0644)
}

func TestSimpleUnmarshal(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	var x = true
	err := Marshal(buf, x, false)
	if err != nil {
		t.Fatal(err)
	}
	//buf.Reset()
	var y bool
	err = Unmarshal(buf, &y)
	if err != nil {
		t.Fatal(err)
	}
	if x != y {
		t.Error("x != y")
	}
}

func TestJexTime(t *testing.T) {
	var t0 = time.Now()
	var j = jexTime(t0)
	var t1 = golangTime(j)
	var d = t1.Sub(t0)
	if d < -time.Microsecond || d > time.Microsecond {
		t.Error("t != t1", t0, t1)
	}
}
