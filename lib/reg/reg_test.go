package reg

import (
	"bytes"
	"math"
	"math/rand"
	"testing"
	"unsafe"
)

var benchSrc []byte
var benchReg *Reg

func randString() string {
	n := 5 + rand.Intn(10)
	buf := bytes.NewBuffer(nil)
	for i := 0; i < n; i++ {
		if rand.Intn(2) == 0 {
			// ascii
			buf.WriteRune(rune(rand.Intn(127)))
		} else {
			// chinese common
			buf.WriteRune(rune(rand.Intn(0x9FFF-0x4E00) + 0x4E00))
		}
	}
	return buf.String()
}

func genBenchObj(depth, width int) interface{} {
	obj := make(map[string]interface{})
	obj["depth"] = float64(depth)
	obj["random float"] = rand.Float64()
	obj["random int"] = float64(rand.Int31())
	obj["nil"] = nil
	obj["true"] = true
	obj["false"] = false
	if depth > 0 {
		var arr []interface{}
		for i := 0; i < width; i++ {
			arr = append(arr, genBenchObj(depth-1, width))
		}
		arr = append(arr, randString())
		obj["array"] = arr
	}
	return obj
}

func initBenchTree() {
	var err error
	benchReg = new(Reg)
	benchReg.j = genBenchObj(4, 4)

	if benchSrc, err = benchReg.Encode(); err != nil {
		panic("generate benchSrc: " + err.Error())
	}

	//ioutil.WriteFile("benchReg.json", benchSrc, 0666)
}

func BenchmarkDecode(b *testing.B) {
	if benchReg == nil {
		b.StopTimer()
		initBenchTree()
		b.StartTimer()
	}
	b.SetBytes(int64(len(benchSrc)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := Decode(benchSrc); err != nil {
				b.Fatal("Decode:", err)
			}
		}
	})
}

func BenchmarkEncode(b *testing.B) {
	if benchSrc == nil {
		b.StopTimer()
		initBenchTree()
		b.StartTimer()
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := benchReg.Encode(); err != nil {
				b.Fatal("Decode:", err)
			}
		}
	})
	b.SetBytes(int64(len(benchSrc)))
}

func TestDecodeJson(t *testing.T) {
	a := `
    {
      "null" : null,
      "bool.true" : true,
      "bool.false" : false,
      "number.123" : 123,
      "number.567.89" : 567.89,
      "string" : "这是个字符串\n\t\"换行缩进\"",
      "array" : [1, 2, 3, "A", "B", "C", { "foo" : "bar"}],
      "object" : {
        "apple" : "苹果",
        "banana" : "香蕉"
      }
    }`
	_, err := Decode([]byte(a))
	if err != nil {
		panic(err)
	}
}

func TestMarshal(t *testing.T) {
	s := new(Reg)
	u1 := uint64(12345678901234)
	if err := s.SetMarshal(`/a/b/c/d/e/f/g/`, u1); err != nil {
		t.Fatal("s.SetMarshal: ", err)
	}
	var u2 uint64
	if err := s.GetUnmarshal(`a/b/c/d/e/f/g`, &u2); err != nil {
		t.Fatal("s.GetUnmarshal: ", err)
	}
	if u2 != u1 {
		t.Fatalf("u1(%d) != u2(%d)", u1, u2)
	}
}

func TestNumbers(t *testing.T) {
	i64s := []int64{
		0, 1, -1,
		math.MinInt32, math.MaxInt32,
		math.MinInt32 - 1, math.MaxInt32 + 1,
		math.MinInt64, math.MaxInt64,
	}
	u64s := []uint64{
		0, 1, math.MaxUint32, math.MaxUint32 + 1, math.MaxUint64,
	}

	rg := new(Reg)
	var err error

	for _, i := range i64s {
		var j int64
		if err = rg.SetInt64("a/b/c", i); err != nil {
			t.Fatal("SetInt64: ", err)
		}
		if j, err = rg.GetInt64("/a/b/c"); err != nil {
			t.Fatal("GetInt64: ", err)
		}
		if j != i {
			t.Fatalf("SetInt64(%d) != GetInt64(%d)", i, j)
		}
	}

	for _, i := range u64s {
		var j uint64
		if err = rg.SetUint64("a//b/c", i); err != nil {
			t.Fatal("SetUint64: ", err)
		}
		if j, err = rg.GetUint64("///a//b///c///"); err != nil {
			t.Fatal("GetUint64: ", err)
		}
		if j != i {
			t.Fatalf("SetUint64(%d) != GetUint64(%d)", i, j)
		}
	}

	if unsafe.Sizeof(*(*int)(nil)) > 4 {
		for _, i := range i64s {
			var j int
			if err = rg.SetInt("a//b//c///", int(i)); err != nil {
				t.Fatal("SetInt: ", err)
			}
			if j, err = rg.GetInt("///a/b/c"); err != nil {
				t.Fatal("GetInt: ", err)
			}
			if j != int(i) {
				t.Fatalf("SetInt(%d) != GetInt(%d)", i, j)
			}
		}

		for _, i := range u64s {
			var j uint
			if err = rg.SetUint("/a/b/c//", uint(i)); err != nil {
				t.Fatal("SetUint: ", err)
			}
			if j, err = rg.GetUint("a/b///c"); err != nil {
				t.Fatal("GetUint: ", err)
			}
			if j != uint(i) {
				t.Fatalf("SetUint(%d) != GetUint(%d)", i, j)
			}
		}
	}
}
