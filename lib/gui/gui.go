// Package gui provide the graphic user interface.
//
// the standard coordinate system used by the gui package, is in pixel unit,
// the origin point (0, 0) is top left of the viewport, positive of x is right,
// positive of y is down, positive of z is toward the inner of screen. just like
// most of other GUI system in the world.
//
// the z axis is typically useless in GUI, but it still a 3D system, can be embeded
// into the 3D scene in some situation.
//
//      |/                 0 ---------- x+
//    - 0 ------- x+       |
//     /|                  |
//    / |                  |
//   /  z+                 |
//  y+                     y+
//
package gui

import (
	"errors"
	"tetra/lib/geom"
	"tetra/lib/glman"
)

//go:generate go run ../../cmd/classp/classp.go .

// Errors
var (
	ErrWrongType = errors.New("Wrong type")
	ErrBadParams = errors.New("Bad params")
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
	// Color type
	Color = glman.Color
)

func init() {
	FactoryRegister()
}
