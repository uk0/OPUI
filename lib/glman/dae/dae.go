package dae

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// COLLADA declares the root of the document that contains some of the content
// in the COLLADA schema.
type COLLADA struct {
	Version string `xml:"version,attr"`
	Asset   *Asset `xml:"asset"`

	LibCameras  *LibCameras  `xml:"library_cameras"`
	LibLights   *LibLights   `xml:"library_lights"`
	LibImages   *LibImages   `xml:"library_images"`
	LibEffects  *LibEffects  `xml:"library_effects"`
	LibMtls     *LibMtls     `xml:"library_materials"`
	LibGeoms    *LibGeoms    `xml:"library_geometries"`
	LibCtrls    *LibCtrls    `xml:"library_controllers"`
	LibVScenes  *LibVScenes  `xml:"library_visual_scenes"`
	LibAniClips *LibAniClips `xml:"library_animation_clips"`
	LibAnis     *LibAnis     `xml:"library_animations"`
	LibFormulas *LibFormulas `xml:"library_formulas"`
	LibNodes    *LibNodes    `xml:"library_nodes"`

	Scene *Scene `xml:"scene"`
}

// Asset defines asset-management information regarding its parent element.
type Asset struct {
	UnitMeter struct {
		Meter float32 `xml:"meter,attr,omitempty"`
		Name  string  `xml:"name,attr,omitempty"`
	} `xml:"unit"`
	UpAxis   string `xml:"up_axis,omitempty"`
	Subject  string `xml:"subject,omitempty"`
	Title    string `xml:"title,omitempty"`
	Revision string `xml:"revision,omitempty"`
	CTime    string `xml:"created,omitempty"`
	MTime    string `xml:"modified,omitempty"`

	Contributors []Contributor `xml:"contributor,omitempty"`
}

// Contributor defines authoring information for asset management.
type Contributor struct {
	Author    string `xml:"author,omitempty"`
	Email     string `xml:"author_email,omitempty"`
	Web       string `xml:"author_website,omitempty"`
	Tool      string `xml:"authoring_tool,omitempty"`
	Copyright string `xml:"copyright,omitempty"`
}

// ValueSid is value with sid
type ValueSid struct {
	Sid string `xml:"sid,attr,omitempty"`
	V   string `xml:",chardata"`
}

// Float parse value as float32
func (f ValueSid) Float() (float32, error) {
	x, err := strconv.ParseFloat(f.V, 32)
	if err != nil {
		return 0, err
	}
	return float32(x), nil
}

// Scene embodies the entire set of information that can be visualized from the contents of a COLLADA resource.
type Scene struct {
	Physics []struct {
		URL string `xml:"url,attr,omitempty"`
	} `xml:"instance_physics_scene"`
	Visual *struct {
		URL string `xml:"url,attr,omitempty"`
	} `xml:"instance_visual_scene"`
	Kinemat *struct {
		URL string `xml:"url,attr,omitempty"`
	} `xml:"instance_kinematics_scene"`
}

// LibCameras provides a library in which to place <camera> elements.
type LibCameras struct {
	ID      string   `xml:"id,attr,omitempty"`
	Name    string   `xml:"name,attr,omitempty"`
	Asset   *Asset   `xml:"asset"`
	Cameras []Camera `xml:"camera"`
	// <imager>  has not technique_common, so not included
}

// Camera declares a view of the visual scene hierarchy or scene graph.
// The camera contains elements that describe the camera’s optics and imager.
type Camera struct {
	ID    string        `xml:"id,attr,omitempty"`
	Name  string        `xml:"name,attr,omitempty"`
	Asset *Asset        `xml:"asset"`
	Persp *Perspective  `xml:"optics>technique_common>perspective"`
	Ortho *Orthographic `xml:"optics>technique_common>orthographic"`
}

// Perspective describes the field of view of a perspective camera.
type Perspective struct {
	XFov   ValueSid `xml:"xfov"`
	YFov   ValueSid `xml:"yfov"`
	Aspect ValueSid `xml:"aspect_ratio"`
	ZNear  ValueSid `xml:"znear"`
	ZFar   ValueSid `xml:"zfar"`
}

// Orthographic describes the field of view of an orthographic camera.
type Orthographic struct {
	XMag   ValueSid `xml:"xmag"`
	YMag   ValueSid `xml:"ymag"`
	Aspect ValueSid `xml:"aspect_ratio"`
	ZNear  ValueSid `xml:"znear"`
	ZFar   ValueSid `xml:"zfar"`
}

// LibLights provides a library in which to place <light> elements.
type LibLights struct {
	ID     string   `xml:"id,attr,omitempty"`
	Name   string   `xml:"name,attr,omitempty"`
	Asset  *Asset   `xml:"asset"`
	Lights []*Light `xml:"light"`
	// <imager>  has not technique_common, so not included
}

// Light declares a light source that illuminates a scene.
type Light struct {
	ID      string      `xml:"id,attr,omitempty"`
	Name    string      `xml:"name,attr,omitempty"`
	Asset   *Asset      `xml:"asset"`
	Ambient *Color      `xml:"technique_common>ambient>color"`
	Direct  *Color      `xml:"technique_common>directional>color"`
	Point   *PointLight `xml:"technique_common>point"`
}

// Color describes the color of its parent light element.
type Color string

// RGB color
func (c Color) RGB() (v [3]float32) {
	fmt.Sscanf(string(c), "%f %f %f", &v[0], &v[1], &v[2])
	return
}

// RGBA color
func (c Color) RGBA() (v [4]float32) {
	fmt.Sscanf(string(c), "%f %f %f %f", &v[0], &v[1], &v[2], &v[3])
	return
}

// PointLight describes a point light source.
type PointLight struct {
	Color  Color     `xml:"color"`
	Const  *ValueSid `xml:"constant_attenuation"`
	Linear *ValueSid `xml:"linear_attenuation"`
	Quad   ValueSid  `xml:"quadratic_attenuation"`
}

// LibImages provides a library for the storage of <image> assets.
type LibImages struct {
	ID     string   `xml:"id,attr,omitempty"`
	Name   string   `xml:"name,attr,omitempty"`
	Asset  *Asset   `xml:"asset"`
	Images []*Image `xml:"image"`
}

// Image declares the storage for the graphical representation of an object.
type Image struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Sid   string `xml:"sid,attr,omitempty"`
	Asset *Asset `xml:"asset"`

	Renderable *struct {
		Share bool `xml:"share,attr"`
	} `xml:"renderable"`

	InitFrom *struct {
		MipsGen bool `xml:"mips_generate,attr"`

		REF    string `xml:"ref,omitempty"` // <init_from><ref>foo.png</ref></init_from>
		ERFOld string `xml:",chardata"`     // <init_from>foo.png</init_from>
		HEX    *struct {
			Format string `xml:"format,attr"`
			Data   string `xml:",chardata"`
		}
	} `xml:"init_from"`

	// TODO: <create_2d>
	// TODO: <create_3d>
	// TODO: <create_cube>
}

// LibEffects provides a library or the storage of <effect> assets.
type LibEffects struct {
	ID      string    `xml:"id,attr,omitempty"`
	Name    string    `xml:"name,attr,omitempty"`
	Asset   *Asset    `xml:"asset"`
	Effects []*Effect `xml:"effect"`
}

// Effect provides a self-contained description of a COLLADA effect.
type Effect struct {
	ID        string         `xml:"id,attr,omitempty"`
	Name      string         `xml:"name,attr,omitempty"`
	Asset     *Asset         `xml:"asset"`
	Annotates []*Annotate    `xml:"annotate"`
	NewParams []*NewParam    `xml:"newparam"`
	Common    *ProfileCommon `xml:"profile_COMMON"`
}

// Annotate adds a strongly typed annotation remark to the parent object.
type Annotate struct {
	Name      string `xml:"name,attr"`
	ValueElem struct {
		XMLName xml.Name
		Data    string `xml:",chardata"`
	} `xml:",any"`
}

// Value parse value_element to go's value
func (a *Annotate) Value() interface{} {
	switch a.ValueElem.XMLName.Local {
	case "string":
		return a.ValueElem.Data
	}
	// TODO: implemenet
	panic("not implemenet")
}

// NewParam creates a new, named parameter object, and assigns it a type and an initial value.
type NewParam struct {
	Sid       string      `xml:"sid,attr,omitempty"`
	Annotates []*Annotate `xml:"annotate"`
	Semantic  string      `xml:"semantic"`
	Modifier  string      `xml:"modifier"` // CONST,UNIFORM,VARYING,STATIC,VOLATILE,EXTERN,SHARED
}

// ProfileCommon opens a block of platform-independent declarations for the common, fixed-function shader.
type ProfileCommon struct {
	ID        string      `xml:"id,attr,omitempty"`
	Asset     *Asset      `xml:"asset"`
	NewParams []*NewParam `xml:"newparam"`
	Technique FxTechnique `xml:"technique"`
}

// FxTechnique Holds a description of the textures, samplers, shaders, parameters,
// and passes necessary for rendering this effect using one method.
type FxTechnique struct {
	ID        string      `xml:"id,attr,omitempty"`
	Sid       string      `xml:"sid,attr,omitempty"`
	Asset     *Asset      `xml:"asset"`
	Annotates []*Annotate `xml:"annotate"`
	// TODO: <blinn>
	// TODO: <constant>
	// TODO: <lambert>
	Phong  *Phong  `xml:"phong"`
	Passes []*Pass `xml:"pass"`
}

// Phong produces a shaded surface where the specular reflection is shaded according the Phong BRDF approximation.
type Phong struct {
	Emission      *FxColorOrTex   `xml:"emission"`
	Ambient       *FxColorOrTex   `xml:"ambient"`
	Diffuse       *FxColorOrTex   `xml:"diffuse"`
	Specular      *FxColorOrTex   `xml:"specular"`
	Shininess     *FxFloatOrParam `xml:"shininess"`
	Reflective    *FxColorOrTex   `xml:"reflective"`
	Reflectivity  *FxFloatOrParam `xml:"reflectivity"`
	Transparent   *FxColorOrTex   `xml:"transparent"`
	Transparency  *FxFloatOrParam `xml:"transparency"`
	IdxRefraction *FxFloatOrParam `xml:"index_of_refraction"`
}

// FxColorOrTex (fx_common_color_or_texture_type) is A type that describes color attributes
// of fixed-function shader elements inside <profile_COMMON> effects.
type FxColorOrTex struct {
	Color   *Color    `xml:"color"`
	Param   *ParamRef `xml:"param"`
	Texture *struct {
		Texture  string `xml:"texture,attr"`
		TexCoord string `xml:"texcoord,attr"`
	} `xml:"texture"`
}

// FxFloatOrParam (fx_common_float_or_param_type) is A type that describes the scalar attributes
// of fixed-function shader elements inside <profile_COMMON> effects.
type FxFloatOrParam struct {
	Float *ValueSid `xml:"float"`
	Param *ParamRef `xml:"param"`
}

// ParamRef references a predefined parameter.
type ParamRef struct {
	ID       string `xml:"id,attr,omitempty"`
	Name     string `xml:"name,attr,omitempty"`
	Semantic string `xml:"semantic,attr,omitempty"`
	Type     string `xml:"type,attr,omitempty"`
	Ref      string `xml:"ref,attr,omitempty"`
}

// Pass provides a static declaration of all the render states, shaders, and settings for one rendering pipeline.
type Pass struct {
	Sid       string      `xml:"sid,attr,omitempty"`
	Annotates []*Annotate `xml:"annotate"`
	// TODO: <states>
	// TODO: <evaluate>
	// TODO:
}

// LibMtls provides a library for the storage of <material> assets.
type LibMtls struct {
	ID    string      `xml:"id,attr,omitempty"`
	Name  string      `xml:"name,attr,omitempty"`
	Asset *Asset      `xml:"asset"`
	Mtls  []*Material `xml:"material"`
}

// Material provides a library for the storage of <material> assets.
type Material struct {
	ID         string     `xml:"id,attr,omitempty"`
	Name       string     `xml:"name,attr,omitempty"`
	Asset      *Asset     `xml:"asset"`
	InstEffect InstEffect `xml:"instance_effect"`
}

// InstEffect instantiates a COLLADA effect.
type InstEffect struct {
	Sid       string      `xml:"sid,attr,omitempty"`
	Name      string      `xml:"name,attr,omitempty"`
	URL       string      `xml:"url,attr,omitempty"`
	TechHints []*TechHint `xml:"technique_hint"`
	SetParam  []*SetParam `xml:"setparam"`
}

// TechHint adds a hint for a platform of which technique to use in this effect.
type TechHint struct {
	Platform string `xml:"platform,attr,omitempty"`
	Ref      string `xml:"ref,attr,omitempty"`
	Profile  string `xml:"profile,attr,omitempty"`
}

// SetParam assigns a new value to a previously defined parameter.
type SetParam struct {
	Ref       string `xml:"ref,attr,omitempty"`
	ValueElem struct {
		XMLName xml.Name
		Data    string `xml:",chardata"`
	} `xml:",any"`
}

// LibGeoms provides a library in which to place <geometry> elements.
type LibGeoms struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}

// LibCtrls provides a library in which to place <controller> elements.
type LibCtrls struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}

// LibVScenes provides a library in which to place <visual_scene> elements.
type LibVScenes struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}

// LibAniClips provides a library in which to place <animation_clip> elements.
type LibAniClips struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}

// LibAnis provides a library in which to place <animation> elements.
type LibAnis struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}

// LibFormulas provides a library in which to place <formula> elements.
type LibFormulas struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}

// LibNodes provides a library in which to place <node> elements.
type LibNodes struct {
	ID    string `xml:"id,attr,omitempty"`
	Name  string `xml:"name,attr,omitempty"`
	Asset *Asset `xml:"asset"`
}
