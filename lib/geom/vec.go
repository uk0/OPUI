package geom

import "math"

// Vec2 is 2D vector
type Vec2 [2]float32

// X component
func (v Vec2) X() float32 { return v[0] }

// Y component
func (v Vec2) Y() float32 { return v[1] }

// Dot product
func (v Vec2) Dot(b Vec2) float32 { return v[0]*b[0] + v[1]*b[1] }

// Kross product
func (v Vec2) Kross(b Vec2) float32 { return v[0]*b[1] - b[0]*v[1] }

// Normal vector
func (v Vec2) Normal() (ret Vec2) {
	a := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1])))
	if a != 0 {
		ret[0], ret[1] = v[0]/a, v[1]/a
		return
	}
	ret = v
	return
}

// Neg returns negative vector
func (v Vec2) Neg() Vec2 {
	return Vec2{-v[0], -v[1]}
}

// Add vector
func (v Vec2) Add(a Vec2) Vec2 {
	return Vec2{v[0] + a[0], v[1] + a[1]}
}

// Sub substract vector
func (v Vec2) Sub(a Vec2) Vec2 {
	return Vec2{v[0] - a[0], v[1] - a[1]}
}

//-------------------------------------------------------------------

// Vec3 is 3D vector
type Vec3 [3]float32

// X component
func (v Vec3) X() float32 { return v[0] }

// Y component
func (v Vec3) Y() float32 { return v[1] }

// Z component
func (v Vec3) Z() float32 { return v[2] }

// Dot product
func (v Vec3) Dot(b Vec3) float32 { return v[0]*b[0] + v[1]*b[1] + v[2]*b[2] }

// Cross product
func (v Vec3) Cross(b Vec3) Vec3 {
	return Vec3{v[1]*b[2] - v[2]*b[1], v[2]*b[0] - v[0]*b[2], v[0]*b[1] - v[1]*b[0]}
}

// Normal vector
func (v Vec3) Normal() (ret Vec3) {
	a := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])))
	if a != 0 {
		ret[0], ret[1], ret[2] = v[0]/a, v[1]/a, v[2]/a
		return
	}
	ret = v
	return
}

// Neg returns negative vector
func (v Vec3) Neg() Vec3 {
	return Vec3{-v[0], -v[1], -v[2]}
}

// Add vector
func (v Vec3) Add(a Vec3) Vec3 {
	return Vec3{v[0] + a[0], v[1] + a[1], v[2] + a[2]}
}

// Sub substract vector
func (v Vec3) Sub(a Vec3) Vec3 {
	return Vec3{v[0] - a[0], v[1] - a[1], v[2] - a[2]}
}

//-------------------------------------------------------------------

// Vec4 is 4D vector
type Vec4 [4]float32

// X component
func (v Vec4) X() float32 { return v[0] }

// Y component
func (v Vec4) Y() float32 { return v[1] }

// Z component
func (v Vec4) Z() float32 { return v[2] }

// W component
func (v Vec4) W() float32 { return v[3] }

// Dot product
func (v Vec4) Dot(b Vec4) float32 { return v[0]*b[0] + v[1]*b[1] + v[2]*b[2] + v[3]*b[3] }

// Normal vector
func (v Vec4) Normal() (ret Vec4) {
	a := float32(math.Sqrt(float64(v[0]*v[0] + v[1]*v[1] + v[2]*v[2] + v[3]*v[3])))
	if a != 0 {
		ret[0], ret[1], ret[2], ret[3] = v[0]/a, v[1]/a, v[2]/a, v[3]/a
		return
	}
	ret = v
	return
}

// Neg returns negative vector
func (v Vec4) Neg() Vec4 {
	return Vec4{-v[0], -v[1], -v[2], -v[3]}
}

// Add vector
func (v Vec4) Add(a Vec4) Vec4 {
	return Vec4{v[0] + a[0], v[1] + a[1], v[2] + a[2], v[3] + a[3]}
}

// Sub substract vector
func (v Vec4) Sub(a Vec4) Vec4 {
	return Vec4{v[0] - a[0], v[1] - a[1], v[2] - a[2], v[3] - a[3]}
}
