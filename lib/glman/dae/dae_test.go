package dae

import (
	"encoding/xml"
	"io/ioutil"
	"testing"
)

func TestParseObj(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/army.dae")
	if err != nil {
		t.Fatal(err)
	}
	var x COLLADA
	err = xml.Unmarshal(data, &x)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("%+v\n", x)
	buf, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(buf))
}
