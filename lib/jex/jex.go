package jex

type Jex struct {
	Any      interface{}
	Children []*Jex
	Index    map[string]*Jex
}
