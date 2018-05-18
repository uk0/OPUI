// Package glman is manage complex 3D objects for render with opengl.
// for example font, text, shapes and particle system.
//
// the standard coordinate system is left-handed, the floor plane is x-y, the above is negtive z.
//
//      |/                 0 ---------- x+
//    - 0 ------- x+       |
//     /|                  |
//    / |                  |
//   /  z+                 |
//  y+                     y+
//
// the standard culling settings is front=CW, back=CCW
//
//   1 -- 3 -- 5 --
//   | \  | \  | \
//   |  \ |  \ |  \
//   0 -- 2 -- 4 --
//
package glman

import (
	"tetra/internal/gl"
	"tetra/lib/geom"
)

type (
	// Rect type
	Rect = geom.Rect
	// Mat4 type
	Mat4 = geom.Mat4
	// Vec2 type
	Vec2 = geom.Vec2
	// Vec3 type
	Vec3 = geom.Vec3
	// Vec4 type
	Vec4 = geom.Vec4
)

var (
	// StackMatP is matrix stack for projection
	StackMatP = geom.Mat4Stack{geom.Mat4Ident()}
	// StackMatV is matrix stack for view/camera
	StackMatV = geom.Mat4Stack{geom.Mat4Ident()}
	// StackMatM is matrix stack for model
	StackMatM = geom.Mat4Stack{geom.Mat4Ident()}
	// StackClip2D is stack for 2D clipping
	StackClip2D = Clip2DStack([]Rect{infClipRect})
)

// GetViewport is convenience wrapper for gl.Get(gl.VIEWPORT)
func GetViewport() Rect {
	var v [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &v[0])
	return Rect{
		float32(v[0]),
		float32(v[1]),
		float32(v[0] + v[2]),
		float32(v[1] + v[3])}
}

// SetViewport is convenience wrapper for gl.Viewport
func SetViewport(rc Rect) {
	gl.Viewport(int32(rc.X0()), int32(rc.Y0()), int32(rc.Width()), int32(rc.Height()))
}
