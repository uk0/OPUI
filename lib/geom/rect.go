package geom

// Rect is 2D rectangle in (x0,y0,x1,y1) format
type Rect [4]float32

// X0 is left position, r[0]
func (r Rect) X0() float32 { return r[0] }

// Y0 is top position, r[1]
func (r Rect) Y0() float32 { return r[1] }

// X1 is right position, r[2]
func (r Rect) X1() float32 { return r[2] }

// Y1 is bottom position, r[3]
func (r Rect) Y1() float32 { return r[3] }

// Center point
func (r Rect) Center() (x, y float32) { return (r[0] + r[2]) * 0.5, (r[1] + r[3]) * 0.5 }

// Width is X1 - X0
func (r Rect) Width() float32 { return r[2] - r[0] }

// Height is Y1 - Y0
func (r Rect) Height() float32 { return r[3] - r[1] }

// IsEmpty reports whether the rect is empty
func (r Rect) IsEmpty() bool { return r[0] == r[2] || r[1] == r[3] }

// Contains reports whether the rect contains the point.
func (r Rect) Contains(x, y float32) bool {
	if r[0] < r[2] {
		if x < r[0] || x > r[2] {
			return false
		}
	} else {
		if x < r[2] || x > r[0] {
			return false
		}
	}
	if r[1] < r[3] {
		if y < r[1] || y > r[3] {
			return false
		}
	} else {
		if y < r[3] || y > r[1] {
			return false
		}
	}
	return true
}

// IsNegative reports x1 < x0 or y1 < y0
func (r Rect) IsNegative() bool { return r[2] < r[0] || r[3] < r[1] }

// PositiveCopy return a copy of r which guarantee x1 >= x0 and y1 >= y0
func (r Rect) PositiveCopy() (x Rect) {
	if r[0] <= r[2] {
		x[0], x[2] = r[0], r[2]
	} else {
		x[2], x[0] = r[0], r[2]
	}
	if r[1] < r[3] {
		x[1], x[3] = r[1], r[3]
	} else {
		x[3], x[1] = r[1], r[3]
	}
	return
}
