package raygun

import (
	"bufio"
	"fmt"
	"math"
)

// EPS is used for the nuances of comparing float values.
const EPS = 1e-9

// http://www.hugi.scene.org/online/coding/hugi%2024%20-%20coding%20graphics%20chris%20dragan%20raytracing%20shapes.htm

// GroupBounds defines a type that has the HitBounds method, such as implemented in Group, Sphere and Plane.
type GroupBounds interface {
	HitBounds(r *Ray) bool
}

// Object is the interface all primitives must have to exist in a raytracing scene.
type Object interface {
	Type() string
	Material() int
	SetMaterial(int)
	Intersect(r *Ray, g, i int) bool
	getNormal(point *Vector) *Vector
	Furthest(point *Vector) float64
	Write(*bufio.Writer)
}

// Base has the default fields and method that is common to all primitives.
type Base struct {
	ObjectType    string
	MaterialIndex int
}

// Type defines the type of the object, it is used when parsing a scene file.
func (b *Base) Type() string {
	return b.ObjectType
}

// Material return the index into the material list defined in the scene.
func (b *Base) Material() int {
	return b.MaterialIndex
}

// SetMaterial set the material index to the supplied value.
func (b *Base) SetMaterial(i int) {
	b.MaterialIndex = i
}

// Sphere is a sphere with position and radius
type Sphere struct {
	Base
	Position *Vector
	Radius   float64
}

// NewSphere creates a new sphere from the supplied values.
func NewSphere(x, y, z, r float64, m int) *Sphere {
	return &Sphere{
		Base: Base{
			ObjectType:    "sphere",
			MaterialIndex: m,
		},
		Position: &Vector{x, y, z},
		Radius:   r,
	}
}

// HitBounds checks if the ray hits this speher.
func (e *Sphere) HitBounds(r *Ray) bool {
	a := r.a
	X := r.origin.Sub(e.Position)
	b := 2 * (r.direction.Dot(X))
	c := X.Dot(X) - e.Radius*e.Radius
	d := b*b - 4*a*c
	if d < 0.0 {
		return false
	}
	disc := math.Sqrt(d)
	t0 := (-b + disc) / 2 * a
	t1 := (-b - disc) / 2 * a
	if t0 < 0.0 && t1 < 0.0 {
		return false
	}
	return true
}

// Intersect calculates if the ray instersect this sphere and at what distance.
func (e *Sphere) Intersect(r *Ray, g, i int) bool {
	a := r.a // r.direction.Dot(r.direction)
	X := r.origin.Sub(e.Position)
	b := 2 * (r.direction.Dot(X))
	c := X.Dot(X) - e.Radius*e.Radius
	d := b*b - 4*a*c
	if d < 0.0 {
		return false
	}
	disc := math.Sqrt(d)
	t0 := (-b + disc) / 2 * a
	t1 := (-b - disc) / 2 * a
	if t0 < 0.0 && t1 < 0.0 {
		return false
	}
	var t float64
	switch {
	case t0 < 0:
		t = t1
	case t1 < 0:
		t = t0
	default:
		t = math.Min(t0, t1)
	}
	if t > r.interDist {
		return false
	}

	r.interDist = t
	r.interObj = i
	r.interGrp = g
	return true
}

func (e *Sphere) getNormal(point *Vector) *Vector {
	normal := point.Sub(e.Position)
	return normal.Normalize()
}

// Furthest calculates the firhest distance this sphere can be from a the point.
// This is used to calculate the group bounds if this sphere is a child of a group.
func (e *Sphere) Furthest(point *Vector) float64 {
	return e.Position.Sub(point).Module() + e.Radius
}

func (e *Sphere) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewSphere(%.2f, %.2f, %.2f, %.2f, %v),\n",
		e.Position.X,
		e.Position.Y,
		e.Position.Z,
		e.Radius,
		e.Material()))
}

// Plane is a primitive that has a position and a normal.
// It has been extended to be a disc - if Radius is set, or a finite plane if width, height
// and depth are set. The nomal need not be along axis.
type Plane struct {
	Base
	Position   *Vector
	Normal     *Vector
	Radius     float64
	Width      float64 // Only two of these will be used for plane, the one in the "normal" direction MUST be 0.0
	Height     float64
	Depth      float64
	HalfWidth  float64
	HalfHeight float64
	HalfDepth  float64
}

func NewPlane(xp, yp, zp, xn, yn, zn, r, w, h, d float64, m int) *Plane {
	p := &Plane{
		Base: Base{
			ObjectType:    "plane",
			MaterialIndex: m,
		},
		Position:   &Vector{xp, yp, zp},
		Normal:     (&Vector{xn, yn, zn}).Normalize(),
		Radius:     r,
		Width:      w,
		Height:     h,
		Depth:      d,
		HalfWidth:  w / 2.0,
		HalfHeight: h / 2.0,
		HalfDepth:  d / 2.0,
	}

	return p
}

func (p *Plane) HitBounds(r *Ray) bool {

	v := p.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}
	t := p.Normal.Dot(p.Position.Sub(r.origin)) / v
	if t < 0.0 {
		return false
	}

	interPoint := r.origin.Add(r.direction.Mul(t))
	if p.HalfWidth > 0.0 && (interPoint.X < p.Position.X-p.HalfWidth || interPoint.X > p.Position.X+p.HalfWidth) {
		return false
	}

	if p.HalfHeight > 0.0 && (interPoint.Y < p.Position.Y-p.HalfHeight || interPoint.Y > p.Position.Y+p.HalfHeight) {
		return false
	}

	if p.HalfDepth > 0.0 && (interPoint.Z < p.Position.Z-p.HalfDepth || interPoint.Z > p.Position.Z+p.HalfDepth) {
		return false
	}

	return true
}

func (p *Plane) Intersect(r *Ray, g, i int) bool {
	v := p.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}
	t := p.Normal.Dot(p.Position.Sub(r.origin)) / v
	if t < 0.0 || t > r.interDist {
		return false
	}
	if p.Radius > 0.0 {
		// We have a disc
		interPoint := r.origin.Add(r.direction.Mul(t))
		dist := interPoint.Sub(p.Position).Module()
		if dist > p.Radius {
			return false
		}
	}

	if p.Width > 0.0 {
		// We have a finite plane
		interPoint := r.origin.Add(r.direction.Mul(t))
		if interPoint.X < p.Position.X-p.HalfWidth || interPoint.X > p.Position.X+p.HalfWidth ||
			interPoint.Y < p.Position.Y-p.HalfHeight || interPoint.Y > p.Position.Y+p.HalfHeight ||
			interPoint.Z < p.Position.Z-p.HalfDepth || interPoint.Z > p.Position.Z+p.HalfDepth {
			return false
		}
	}

	r.interDist = t
	r.interObj = i
	r.interGrp = g
	return true

}

func (p *Plane) getNormal(point *Vector) *Vector {
	return p.Normal
}

func (p *Plane) Furthest(point *Vector) float64 {
	return p.Position.Sub(point).Module() + p.Radius + p.Width/2.0
}

// func (p *Plane) String() string {
// 	return fmt.Sprintf("<Pla: %d %s %.2f>", p.Material, p.Normal.String(), p.Radius)
// }

func (p *Plane) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewPlane(%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %v),\n",
		p.Position.X,
		p.Position.Y,
		p.Position.Z,
		p.Normal.X,
		p.Normal.Y,
		p.Normal.Z,
		p.Radius,
		p.Width,
		p.Material()))
}

// Cube

type Cube struct {
	Base
	Position *Vector
	Width    float64
	Height   float64
	Depth    float64
	Min      *Vector
	Max      *Vector
}

func NewCube(x, y, z, w, h, d float64, m int) *Cube {
	c := &Cube{
		Base: Base{
			ObjectType:    "cube",
			MaterialIndex: m,
		},
		Position: &Vector{x, y, z},
		Width:    w, // x direction
		Height:   h, // y direction
		Depth:    d, // z direction
	}
	c.initMinMax()
	return c
}

func (c *Cube) Intersect(r *Ray, g, i int) bool {
	n := c.Min.Sub(r.origin).Div(r.direction)
	f := c.Max.Sub(r.origin).Div(r.direction)
	n, f = n.Min(f), n.Max(f)
	t0 := math.Max(math.Max(n.X, n.Y), n.Z)
	t1 := math.Min(math.Min(f.X, f.Y), f.Z)
	if t0 > 0 && t0 < t1 {
		if t0 > r.interDist {
			return false
		}
		r.interDist = t0
		r.interObj = i
		r.interGrp = g
		return true
	}
	return false
}

func (c *Cube) getNormal(point *Vector) *Vector {

	switch {
	case point.X < c.Min.X+EPS:
		return &Vector{-1, 0, 0}
	case point.X > c.Max.X-EPS:
		return &Vector{1, 0, 0}
	case point.Y < c.Min.Y+EPS:
		return &Vector{0, -1, 0}
	case point.Y > c.Max.Y-EPS:
		return &Vector{0, 1, 0}
	case point.Z < c.Min.Z+EPS:
		return &Vector{0, 0, -1}
	case point.Z > c.Max.Z-EPS:
		return &Vector{0, 0, 1}
	}
	return &Vector{0, 1, 0}
}

func (c *Cube) initMinMax() {
	c.Min = &Vector{
		c.Position.X - c.Width/2.0,
		c.Position.Y - c.Height/2.0,
		c.Position.Z,
	}
	c.Max = &Vector{
		c.Position.X + c.Width/2.0,
		c.Position.Y + c.Height/2.0,
		c.Position.Z + c.Depth,
	}
}

func (c *Cube) Furthest(point *Vector) float64 {
	max := math.Max(math.Max(c.Width/2.0, c.Height/2.), c.Depth)
	return c.Position.Sub(point).Module() + 1.5*max
}

func (c *Cube) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewCube(%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %v),\n",
		c.Position.X,
		c.Position.Y,
		c.Position.Z,
		c.Width,
		c.Height,
		c.Depth,
		c.Material(),
	))
}

type Cylinder struct {
	Base
	Position  *Vector
	Direction *Vector
	Length    float64
	Radius    float64
	StartDisc *Plane
	EndDisc   *Plane
}

func NewCylinder(xp, yp, zp, xd, yd, zd, l, r float64, m int) *Cylinder {
	c := &Cylinder{
		Base: Base{
			ObjectType:    "cylinder",
			MaterialIndex: m,
		},
		Position:  &Vector{xp, yp, zp},
		Direction: (&Vector{xd, yd, zd}).Normalize(),
		Length:    l,
		Radius:    r,
	}
	pos := c.Position
	dir := c.Direction.Mul(-1)
	c.StartDisc = NewPlane(pos.X, pos.Y, pos.Z, dir.X, dir.Y, dir.Z, c.Radius, 0.0, 0.0, 0.0, c.Material())

	pos = c.Position.Add(c.Direction.Mul(c.Length))
	dir = c.Direction
	c.EndDisc = NewPlane(pos.X, pos.Y, pos.Z, dir.X, dir.Y, dir.Z, c.Radius, 0.0, 0.0, 0.0, c.Material())
	return c
}

// http://blog.makingartstudios.com/?p=286
func (y *Cylinder) Intersect(r *Ray, g, i int) bool {
	cylend := y.Position.Add(y.Direction.Mul(y.Length))
	AB := cylend.Sub(y.Position)
	AO := r.origin.Sub(y.Position)

	ABDotD := AB.Dot(r.direction)
	ABDotAO := AB.Dot(AO)
	ABDotAB := AB.Dot(AB)

	m := ABDotD / ABDotAB
	n := ABDotAO / ABDotAB

	Q := r.direction.Sub(AB.Mul(m))
	R := AO.Sub(AB.Mul(n))

	a := Q.Dot(Q)
	b := 2.0 * Q.Dot(R)
	c := R.Dot(R) - y.Radius*y.Radius

	if a == 0.0 {
		return false
	}

	d := b*b - 4.0*a*c
	if d < 0.0 {
		return false
	}

	t0 := (-b - math.Sqrt(d)) / (2 * a)
	t1 := (-b + math.Sqrt(d)) / (2 * a)

	if t0 < 0.0 && t1 < 0.0 {
		return false
	}

	if t0 > t1 {
		t0, t1 = t1, t0
	}

	var t float64
	if t0 < 0.0 {
		t = t1
	} else {
		t = math.Min(t0, t1)
	}

	if t > r.interDist {
		return false
	}

	tk := t*m + n
	if tk < 0.0 {
		// Could be on start cap
		if y.intersectCap(r, true) && t1 > r.interDist {
			t = t1
		} else {
			return false // y.intersectCap(r, i, true)
		}
	}

	if tk > 1.0 {
		// Could be on end cap
		if y.intersectCap(r, false) && t1 > r.interDist {
			t = t1
		} else {
			return false // y.intersectCap(r, i, true)
		}
	}

	r.interDist = t
	r.interObj = i
	r.interGrp = g
	return true

}
func (y *Cylinder) intersectCap(r *Ray, start bool) bool {
	if start {
		return y.StartDisc.Intersect(r, 0, 0)
	}
	return y.EndDisc.Intersect(r, 0, 0)
}

func (y *Cylinder) getNormal(point *Vector) *Vector {
	PQ := point.Sub(y.Position)
	pqa := PQ.Dot(y.Direction)
	PQAA := y.Direction.Mul(pqa)
	return PQ.Sub(PQAA).Normalize()
}

func (y *Cylinder) Furthest(point *Vector) float64 {
	return y.Position.Sub(point).Module() + y.Length + y.Radius
}

func (y *Cylinder) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewCylinder(%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %v),\n",
		y.Position.X,
		y.Position.Y,
		y.Position.Z,
		y.Direction.X,
		y.Direction.Y,
		y.Direction.Z,
		y.Length,
		y.Radius,
		y.Material(),
	))
}
