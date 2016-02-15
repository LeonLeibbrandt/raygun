package raygun

import (
	"fmt"
	"math"
)

const EPS = 1e-9

// http://www.hugi.scene.org/online/coding/hugi%2024%20-%20coding%20graphics%20chris%20dragan%20raytracing%20shapes.htm

// Object
type Object interface {
	Type() string
	MaterialIndex() int
	Intersect(r *Ray, i int) bool
	getNormal(point *Vector) *Vector
}

// Sphere
type Sphere struct {
	Material int
	Position *Vector
	Radius   float64
}

func NewSphere(x, y, z, r float64, m int) *Sphere {
	return &Sphere{
		Material: m,
		Position: &Vector{x, y, z},
		Radius:   r,
	}
}

func (e *Sphere) Type() string {
	return "sphere"
}

func (e *Sphere) MaterialIndex() int {
	return e.Material
}

func (e *Sphere) Intersect(r *Ray, i int) bool {
	a := r.direction.Dot(r.direction)
	X := r.origin.Sub(e.Position)
	b := 2 * (r.direction.Dot(X))
	c := X.Dot(X) - e.Radius*e.Radius
	if b*b < 4*a*c {
		return false
	}
	disc := math.Sqrt(b*b - 4*a*c)
	t0 := (-b + disc) / 2 * a
	t1 := (-b - disc) / 2 * a
	if t0 < 0.0 && t1 < 0.0 {
		return false
	}
	var d float64
	switch {
	case t0 < 0:
		d = t1
	case t1 < 0:
		d = t0
	default:
		d = math.Min(t0, t1)
	}
	if d > r.interDist {
		return false
	}

	r.interDist = d
	r.interObj = i
	return true
}

func (e *Sphere) getNormal(point *Vector) *Vector {
	normal := point.Sub(e.Position)
	return normal.Normalize()
}

func (e *Sphere) String() string {
	return fmt.Sprintf("<Esf: %d %s %.2f>", e.Material, e.Position.String(), e.Radius)
}

// PLANE
type Plane struct {
	Material int
	Position *Vector
	Normal   *Vector
	distance float64
}

func NewPlane(xp, yp, zp, xn, yn, zn, d float64, m int) *Plane {
	return &Plane{
		Material: m,
		Position: &Vector{xp, yp, zp},
		Normal:   (&Vector{xn, yn, zn}).Normalize(),
		distance: d,
	}
}

func (p *Plane) Type() string {
	return "plane"
}

func (p *Plane) MaterialIndex() int {
	return p.Material
}

func (p *Plane) Intersect(r *Ray, i int) bool {
	v := p.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}
	t := p.Normal.Dot(p.Position.Sub(r.origin)) / v
	if t < 0.0 || t > r.interDist {
		return false
	}
	if p.distance > 0.0 {
		// We have a disc
		interPoint := r.origin.Add(r.direction.Mul(t))
		dist := interPoint.Sub(p.Position).Module()
		if dist > p.distance {
			return false
		}
	}

	r.interDist = t
	r.interObj = i
	return true

}

func (p *Plane) getNormal(point *Vector) *Vector {
	return p.Normal
}

func (p *Plane) String() string {
	return fmt.Sprintf("<Pla: %d %s %.2f>", p.Material, p.Normal.String(), p.distance)
}

// Cube

type Cube struct {
	Material int
	Position *Vector
	Width    float64
	Height   float64
	Depth    float64
	min      *Vector
	max      *Vector
}

func NewCube(x, y, z, w, h, d float64, m int) *Cube {
	c := &Cube{
		Material: m,
		Position: &Vector{x, y, z},
		Width:    w, // x direction
		Height:   h, // y direction
		Depth:    d, // z direction
	}
	c.initMinMax()
	return c
}

func (c *Cube) Type() string {
	return "cube"
}

func (c *Cube) MaterialIndex() int {
	return c.Material
}

func (c *Cube) Intersect(r *Ray, i int) bool {
	n := c.min.Sub(r.origin).Div(r.direction)
	f := c.max.Sub(r.origin).Div(r.direction)
	n, f = n.Min(f), n.Max(f)
	t0 := math.Max(math.Max(n.x, n.y), n.z)
	t1 := math.Min(math.Min(f.x, f.y), f.z)
	if t0 > 0 && t0 < t1 {
		if t0 > r.interDist {
			return false
		}
		r.interDist = t0
		r.interObj = i
		return true
	}
	return false
}

func (c *Cube) getNormal(point *Vector) *Vector {

	switch {
	case point.x < c.min.x+EPS:
		return &Vector{-1, 0, 0}
	case point.x > c.max.x-EPS:
		return &Vector{1, 0, 0}
	case point.y < c.min.y+EPS:
		return &Vector{0, -1, 0}
	case point.y > c.max.y-EPS:
		return &Vector{0, 1, 0}
	case point.z < c.min.z+EPS:
		return &Vector{0, 0, -1}
	case point.z > c.max.z-EPS:
		return &Vector{0, 0, 1}
	}
	return &Vector{0, 1, 0}
}

func (c *Cube) initMinMax() {
	c.min = &Vector{
		c.Position.x - c.Width/2.0,
		c.Position.y - c.Height/2.0,
		c.Position.z,
	}
	c.max = &Vector{
		c.Position.x + c.Width/2.0,
		c.Position.y + c.Height/2.0,
		c.Position.z + c.Depth,
	}
}

type Cylinder struct {
	Material  int
	Position  *Vector
	Direction *Vector
	Length    float64
	Radius    float64
}

func NewCylinder(xp, yp, zp, xd, yd, zd, l, r float64, m int) *Cylinder {
	return &Cylinder{
		Material:  m,
		Position:  &Vector{xp, yp, zp},
		Direction: (&Vector{xd, yd, zd}).Normalize(),
		Length:    l,
		Radius:    r,
	}
}

func (y *Cylinder) Type() string {
	return "cylinder"
}

func (y *Cylinder) MaterialIndex() int {
	return y.Material
}

// http://blog.makingartstudios.com/?p=286
func (y *Cylinder) Intersect(r *Ray, i int) bool {
	cylend := y.Position.Add(y.Direction.Mul(y.Length))
	AB := cylend.Sub(y.Position)
	AO := r.origin.Sub(y.Position)

	AB_dot_d := AB.Dot(r.direction)
	AB_dot_AO := AB.Dot(AO)
	AB_dot_AB := AB.Dot(AB)

	m := AB_dot_d / AB_dot_AB
	n := AB_dot_AO / AB_dot_AB

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

	t_k := t*m + n
	if t_k < 0.0 {
		// Could be on start cap
		if y.intersectCap(r, i, true) && t1 > r.interDist {
			t = t1
		} else {
			return false // y.intersectCap(r, i, true)
		}
	}

	if t_k > 1.0 {
		// Could be on end cap
		if y.intersectCap(r, i, false) && t1 > r.interDist {
			t = t1
		} else {
			return false // y.intersectCap(r, i, true)
		}
	}

	r.interDist = t
	r.interObj = i
	return true

}
func (y *Cylinder) intersectCap(r *Ray, i int, start bool) bool {
	var pos *Vector
	var dir *Vector
	if start {
		pos = y.Position
		dir = y.Direction.Mul(-1)
	} else {
		pos = y.Position.Add(y.Direction.Mul(y.Length))
		dir = y.Direction
	}

	end := Plane{y.Material, pos, dir, y.Radius}
	return end.Intersect(r, i)
}

func (y *Cylinder) getNormal(point *Vector) *Vector {
	PQ := point.Sub(y.Position)
	pqa := PQ.Dot(y.Direction)
	PQAA := y.Direction.Mul(pqa)
	return PQ.Sub(PQAA).Normalize()
}
