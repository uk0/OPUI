package skin

// Fallback is a minimal skin, simple but just works,
// just new it then use, no config file, no other resource.
type Fallback struct {
	Common
}

// Init the object
func (sk *Fallback) Init() {
	sk.Common.Init()
}
