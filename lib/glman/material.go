package glman

var (
	// MtlIndex is index of materials
	MtlIndex = make(map[string]*MtlInfo)
)

// MtlInfo is surface material
type MtlInfo struct {
	Diffuse       [3]float32
	Ambient       [3]float32
	Specular      [3]float32
	Emissive      [3]float32
	SpecularPower float32
	Alpha         float32
	RenderMode    uint32
	MapDiffuse    string
	MapSpecular   string
	MapBump       string
	MapAlpha      string
}

// Mtl is material loaded into memory
type Mtl struct {
	MtlInfo
	TexMapDiffuse  *Res
	TexMapSpecular *Res
	TexMapBump     *Res
	TexMapAlpha    *Res
}

// LoadMtl load material
func LoadMtl(name string) *Mtl {
	info := MtlIndex[name]
	if info == nil {
		return nil
	}
	p := &Mtl{MtlInfo: *info}
	p.TexMapDiffuse = LoadTexture(p.MapDiffuse)
	p.TexMapSpecular = LoadTexture(p.MapSpecular)
	p.TexMapBump = LoadTexture(p.MapBump)
	p.TexMapAlpha = LoadTexture(p.MapAlpha)
	return p
}
