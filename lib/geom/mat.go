package geom

import "math"

// Mat4 is 4x4 matrix
type Mat4 [16]float32

var ident4 = Mat4{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}

// Mat4Ident return identity
func Mat4Ident() Mat4 {
	return ident4
}

// Mult calculate  a * b
func (a Mat4) Mult(b Mat4) (c Mat4) {
	c[0] = a[0]*b[0] + a[4]*b[1] + a[8]*b[2] + a[12]*b[3]
	c[1] = a[1]*b[0] + a[5]*b[1] + a[9]*b[2] + a[13]*b[3]
	c[2] = a[2]*b[0] + a[6]*b[1] + a[10]*b[2] + a[14]*b[3]
	c[3] = a[3]*b[0] + a[7]*b[1] + a[11]*b[2] + a[15]*b[3]
	c[4] = a[0]*b[4] + a[4]*b[5] + a[8]*b[6] + a[12]*b[7]
	c[5] = a[1]*b[4] + a[5]*b[5] + a[9]*b[6] + a[13]*b[7]
	c[6] = a[2]*b[4] + a[6]*b[5] + a[10]*b[6] + a[14]*b[7]
	c[7] = a[3]*b[4] + a[7]*b[5] + a[11]*b[6] + a[15]*b[7]
	c[8] = a[0]*b[8] + a[4]*b[9] + a[8]*b[10] + a[12]*b[11]
	c[9] = a[1]*b[8] + a[5]*b[9] + a[9]*b[10] + a[13]*b[11]
	c[10] = a[2]*b[8] + a[6]*b[9] + a[10]*b[10] + a[14]*b[11]
	c[11] = a[3]*b[8] + a[7]*b[9] + a[11]*b[10] + a[15]*b[11]
	c[12] = a[0]*b[12] + a[4]*b[13] + a[8]*b[14] + a[12]*b[15]
	c[13] = a[1]*b[12] + a[5]*b[13] + a[9]*b[14] + a[13]*b[15]
	c[14] = a[2]*b[12] + a[6]*b[13] + a[10]*b[14] + a[14]*b[15]
	c[15] = a[3]*b[12] + a[7]*b[13] + a[11]*b[14] + a[15]*b[15]
	return
}

func (m Mat4) MultVec4(a [4]float32) (b [4]float32) {
	b[0] = a[0]*m[0] + a[1]*m[4] + a[2]*m[8] + a[3]*m[12]
	b[1] = a[0]*m[1] + a[1]*m[5] + a[2]*m[9] + a[3]*m[13]
	b[2] = a[0]*m[2] + a[1]*m[6] + a[2]*m[10] + a[3]*m[14]
	b[3] = a[0]*m[3] + a[1]*m[7] + a[2]*m[11] + a[3]*m[15]
	return
}

func (m Mat4) MultVec3(a Vec3) (b Vec3) {
	tmp := Vec4{a[0], a[1], a[2], 1}
	tmp = m.MultVec4(tmp)
	b = Vec3{tmp[0] / tmp[3], tmp[1] / tmp[3], tmp[2] / tmp[3]}
	return
}

func (m Mat4) Transpose() (x Mat4) {
	x = m
	x[1], x[4] = x[4], x[1]
	x[2], x[8] = x[8], x[2]
	x[3], x[12] = x[12], x[3]
	x[6], x[9] = x[9], x[6]
	x[7], x[13] = x[13], x[7]
	x[11], x[14] = x[14], x[11]
	return
}

func sincos(rad float32) (s, c float32) {
	a, b := math.Sincos(float64(rad))
	s, c = float32(a), float32(b)
	return
}

func Mat4RotX(rad float32) (rot Mat4) {
	sin_a, cos_a := sincos(-rad)
	rot = ident4
	rot[5] = cos_a
	rot[9] = sin_a
	rot[6] = -sin_a
	rot[10] = cos_a
	return
}

func Mat4RotY(rad float32) (rot Mat4) {
	sin_a, cos_a := sincos(-rad)
	rot = ident4
	rot[0] = cos_a
	rot[8] = -sin_a
	rot[2] = sin_a
	rot[10] = cos_a
	return
}

func Mat4RotZ(rad float32) (rot Mat4) {
	sin_a, cos_a := sincos(-rad)
	rot = ident4
	rot[0] = cos_a
	rot[4] = sin_a
	rot[1] = -sin_a
	rot[5] = cos_a
	return
}

func Mat4Trans(_dx float32, _dy float32, _dz float32) (trans Mat4) {
	trans = ident4
	trans[12] = _dx
	trans[13] = _dy
	trans[14] = _dz
	return
}

func Mat4Scale(_sx float32, _sy float32, _sz float32) (n Mat4) {
	n = ident4
	n[0] = _sx
	n[5] = _sy
	n[10] = _sz
	return
}

func Mat4Rot(rad float32, _x float32, _y float32, _z float32) (rot Mat4) {
	u := Vec3{_x, _y, _z}
	u = u.Normal()

	ux := u[0]
	uy := u[1]
	uz := u[2]

	ux2 := u[0] * u[0]
	uy2 := u[1] * u[1]
	uz2 := u[2] * u[2]

	uxy := u[0] * u[1]
	uyz := u[1] * u[2]
	uxz := u[0] * u[2]

	s, c := sincos(rad)

	c1 := 1 - c

	rot = ident4

	rot[0] = ux2 + (1-ux2)*c
	rot[4] = uxy*c1 - uz*s
	rot[8] = uxz*c1 + uy*s

	rot[1] = uxy*c1 + uz*s
	rot[5] = uy2 + (1-uy2)*c
	rot[9] = uyz*c1 - ux*s

	rot[2] = uxz*c1 - uy*s
	rot[6] = uyz*c1 + ux*s
	rot[10] = uz2 + (1-uz2)*c

	return
}

func (m Mat4) Det() float32 {
	s0 := m[0]*m[5] - m[4]*m[1]
	s1 := m[0]*m[6] - m[4]*m[2]
	s2 := m[0]*m[7] - m[4]*m[3]
	s3 := m[1]*m[6] - m[5]*m[2]
	s4 := m[1]*m[7] - m[5]*m[3]
	s5 := m[2]*m[7] - m[6]*m[3]

	c5 := m[10]*m[15] - m[14]*m[11]
	c4 := m[9]*m[15] - m[13]*m[11]
	c3 := m[9]*m[14] - m[13]*m[10]
	c2 := m[8]*m[15] - m[12]*m[11]
	c1 := m[8]*m[14] - m[12]*m[10]
	c0 := m[8]*m[13] - m[12]*m[9]

	det := s0*c5 - s1*c4 + s2*c3 + s3*c2 - s4*c1 + s5*c0
	return det
}

func (m Mat4) Inverse() (b Mat4) {
	s0 := m[0]*m[5] - m[4]*m[1]
	s1 := m[0]*m[6] - m[4]*m[2]
	s2 := m[0]*m[7] - m[4]*m[3]
	s3 := m[1]*m[6] - m[5]*m[2]
	s4 := m[1]*m[7] - m[5]*m[3]
	s5 := m[2]*m[7] - m[6]*m[3]

	c5 := m[10]*m[15] - m[14]*m[11]
	c4 := m[9]*m[15] - m[13]*m[11]
	c3 := m[9]*m[14] - m[13]*m[10]
	c2 := m[8]*m[15] - m[12]*m[11]
	c1 := m[8]*m[14] - m[12]*m[10]
	c0 := m[8]*m[13] - m[12]*m[9]

	det := s0*c5 - s1*c4 + s2*c3 + s3*c2 - s4*c1 + s5*c0
	if det == 0 {
		// panic("Can not calc inverse Mat4, det == 0")
		return ident4
	}

	invdet := 1.0 / det

	b[0] = (m[5]*c5 - m[6]*c4 + m[7]*c3) * invdet
	b[1] = (-m[1]*c5 + m[2]*c4 - m[3]*c3) * invdet
	b[2] = (m[13]*s5 - m[14]*s4 + m[15]*s3) * invdet
	b[3] = (-m[9]*s5 + m[10]*s4 - m[11]*s3) * invdet

	b[4] = (-m[4]*c5 + m[6]*c2 - m[7]*c1) * invdet
	b[5] = (m[0]*c5 - m[2]*c2 + m[3]*c1) * invdet
	b[6] = (-m[12]*s5 + m[14]*s2 - m[15]*s1) * invdet
	b[7] = (m[8]*s5 - m[10]*s2 + m[11]*s1) * invdet

	b[8] = (m[4]*c4 - m[5]*c2 + m[7]*c0) * invdet
	b[9] = (-m[0]*c4 + m[1]*c2 - m[3]*c0) * invdet
	b[10] = (m[12]*s4 - m[13]*s2 + m[15]*s0) * invdet
	b[11] = (-m[8]*s4 + m[9]*s2 - m[11]*s0) * invdet

	b[12] = (-m[4]*c3 + m[5]*c1 - m[6]*c0) * invdet
	b[13] = (m[0]*c3 - m[1]*c1 + m[2]*c0) * invdet
	b[14] = (-m[12]*s3 + m[13]*s1 - m[14]*s0) * invdet
	b[15] = (m[8]*s3 - m[9]*s1 + m[10]*s0) * invdet

	return b
}

func Mat4Ortho(xMin float32, xMax float32, yMin float32, yMax float32, zMin float32, zMax float32) (m Mat4) {
	w := xMax - xMin
	h := yMin - yMax
	d := zMax - zMin

	m = ident4

	// (l', r') => (-1, 1)
	m[0] = 2 / w

	// (t', b') => (1, -1)
	m[5] = -2 / h

	// (n, f) => (0, 1)
	m[10] = -1 / d
	m[14] = 0.5
	return
}

func Mat4Frustum(_left float32, _right float32, _bottom float32, _top float32, _near float32, _far float32) (m Mat4) {
	m = ident4

	//    left        right
	//      +----------+
	//
	//
	//  + eye

	m[0] = 2 * _near / (_right - _left)
	m[5] = 2 * _near / (_top - _bottom)
	m[8] = (_right + _left) / (_right - _left)
	m[9] = (_top + _bottom) / (_top - _bottom)
	m[10] = -(_far + _near) / (_far - _near)
	m[11] = -1
	m[14] = -2 * _far * _near / (_far - _near)
	return
}

func Mat4Perspective(_fovy float32, _aspect float32, _near float32, _far float32) (m Mat4) {
	//                                    |         |
	//    \                     /         +         +-- m32
	// f  -+-------------------+          |         |
	//      \                 / \         |         |
	//       \    viewing    /   +---> ---+---------+--  1
	//        \   frustum   /             |         |
	//         \           /              |         |
	// m22 -----+         +-------- -> ---+         +--  0.5
	//           \       /                |         |
	//            \     /                 |         |
	// n  ---------+---+----------- -> ---+=========+--  0
	//              \ /                   |         |
	//               +                    |         |

	// x => [-1, +1],  y => [-1, +1], z => [0, +1]

	n := _near
	f := _far
	e := 1 / float32(math.Tan(float64(_fovy)/2)) // focus length
	m22 := f / (n - f)
	m32 := n * f / (n - f)

	m = ident4
	m[0] = e
	m[5] = e / _aspect
	m[10] = m22
	m[11] = -1
	m[14] = m32
	m[15] = 0
	return
}

func Mat4LookAt(_eye Vec3, _lookAt Vec3, _up Vec3) (m Mat4) {
	z := _eye.Sub(_lookAt).Normal()
	x := _up.Cross(z).Normal()
	y := z.Cross(x)

	m = [16]float32{
		x[0], y[0], z[0], 0,
		x[1], y[1], z[1], 0,
		x[2], y[2], z[2], 0,
		-x.Dot(_eye), -y.Dot(_eye), -z.Dot(_eye), 1,
	}

	return
}
