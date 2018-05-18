package factory

import "log"

var (
	ctors = make(map[string]func() interface{})
)

// Register factory method for class
func Register(class string, fn func() interface{}) {
	// technically all strings are acceptable, we just void the most common mistake.
	if class == "" || class[0] <= ' ' || class[len(class)-1] <= ' ' {
		log.Panicf("invalid class \"%s\".\n", class)
	}
	if _, exist := ctors[class]; exist {
		log.Printf("warning: override class \"%s\".\n", class)
	}
	ctors[class] = fn
}

// New a object with creator, then call FactoryAsm method if exist.
func New(class string) interface{} {
	ctor, ok := ctors[class]
	if !ok {
		log.Panicf("\"%s\" not found.\n", class)
	}
	return ctor()
}

// Get the factory method for class
func Get(class string) func() interface{} {
	return ctors[class]
}
