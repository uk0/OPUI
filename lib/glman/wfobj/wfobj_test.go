package wfobj

import (
	"os"
	"testing"
)

func TestParseMtl(t *testing.T) {
	f, err := os.Open("testdata/rock.mtl")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := ParseMtl(f, "rock.mtl")
	if err != nil {
		t.Fatal(err)
	}
	if m["Material"].MapDiffuse != "Rock-Texture-Surface.jpg" {
		t.Fatalf("Material.MapDiffuse = \"%s\", want \"%s\"",
			m["Material"].MapDiffuse, "Rock-Texture-Surface.jpg")
	}

	// for k, x := range m {
	// 	t.Logf("%s:\n%+v\n", k, x)
	// }
}

func TestParseObj(t *testing.T) {
	f, err := os.Open("testdata/rock.obj")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := ParseObj(f, "rock.obj")
	if err != nil {
		t.Fatal(err)
	}
	m.Name = "rock"
	t.Logf("%v\n", m)
	if len(m.Groups) != 2 ||
		len(m.Groups[0].V) != 12 ||
		len(m.Groups[1].V) != 1920 ||
		m.Groups[0].Material != "Material.001" ||
		m.Groups[1].Material != "Material" {
		t.Fatal("incorrect result")
	}
}
