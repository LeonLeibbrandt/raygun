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
	Material() int
	Intersect(r *Ray, i int) bool
	getNormal(point Vector) Vector
}

// Sphere
type Sphere struct {
	material int
	position Vector
	radius   float64
}

func NewSphere(x, y, z, r float64, m int) *Sphere {
	return &Sphere{
		material: m,
		position: Vector{x, y, z},
		radius:   r,
	}
}

func (e Sphere) Type() string {
	return "sphere"
}

func (e Sphere) Material() int {
	return e.material
}

func (e Sphere) Intersect(r *Ray, i int) bool {
	a := r.direction.Dot(r.direction)
	X := r.origin.Sub(e.position)
	b := 2 * (r.direction.Dot(X))
	c := X.Dot(X) - e.radius*e.radius
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

func (e Sphere) getNormal(point Vector) Vector {
	normal := point.Sub(e.position)
	return normal.Normalize()
}

func (e Sphere) String() string {
	return fmt.Sprintf("<Esf: %d %s %.2f>", e.material, e.position.String(), e.radius)
}

// PLANE
type Plane struct {
	material int
	position Vector
	normal   Vector
	distance float64
}

func NewPlane(xp, yp, zp, xn, yn, zn, d float64, m int) *Plane {
	return &Plane{
		material: m,
		position: Vector{xp, yp, zp},
		normal:   Vector{xn, yn, zn}.Normalize(),
		distance: d,
	}
}

func (p Plane) Type() string {
	return "plane"
}

func (p Plane) Material() int {
	return p.material
}

func (p Plane) Intersect(r *Ray, i int) bool {
	v := p.normal.Dot(r.direction)
	if v == 0 {
		return false
	}
	t := p.normal.Dot(p.position.Sub(r.origin)) / v
	if t < 0.0 || t > r.interDist {
		return false
	}
	if p.distance > 0.0 {
		// We have a disc
		interPoint := r.origin.Add(r.direction.Mul(t))
		dist := interPoint.Sub(p.position).Module()
		if dist > p.distance {
			return false
		}
		// fmt.Printf("%v %v %v %v %v\n", interPoint, dist, p.distance, r.interDist, t)
	}

	r.interDist = t
	r.interObj = i
	return true

}

func (p Plane) getNormal(point Vector) Vector {
	return p.normal
}

func (p Plane) String() string {
	return fmt.Sprintf("<Pla: %d %s %.2f>", p.material, p.normal.String(), p.distance)
}

// Cube

type Cube struct {
	material int
	min      Vector
	max      Vector
}

func NewCube(x, y, z, w, h, d float64, m int) *Cube {
	return &Cube{
		material: m,
		min:      Vector{x - w/2.0, y - h/2.0, z - d/2.0},
		max:      Vector{x + w/2.0, y + h/2.0, z + d/2.0},
	}

}

func (c Cube) Type() string {
	return "cube"
}

func (c Cube) Material() int {
	return c.material
}

func (c Cube) Intersect(r *Ray, i int) bool {
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

func (c Cube) getNormal(p Vector) Vector {
	switch {
	case p.x < c.min.x+EPS:
		return Vector{-1, 0, 0}
	case p.x > c.max.x-EPS:
		return Vector{1, 0, 0}
	case p.y < c.min.y+EPS:
		return Vector{0, -1, 0}
	case p.y > c.max.y-EPS:
		return Vector{0, 1, 0}
	case p.z < c.min.z+EPS:
		return Vector{0, 0, -1}
	case p.z > c.max.z-EPS:
		return Vector{0, 0, 1}
	}
	return Vector{0, 1, 0}
}

type Cylinder struct {
	material  int
	position  Vector
	direction Vector
	length    float64
	radius    float64
}

func NewCylinder(xp, yp, zp, xd, yd, zd, l, r float64, m int) *Cylinder {
	return &Cylinder{
		material:  m,
		position:  Vector{xp, yp, zp},
		direction: Vector{xd, yd, zd}.Normalize(),
		length:    l,
		radius:    r,
	}
}

func (y Cylinder) Type() string {
	return "cylinder"
}

func (y Cylinder) Material() int {
	return y.material
}

// http://blog.makingartstudios.com/?p=286
func (y Cylinder) Intersect(r *Ray, i int) bool {
	cylend := y.position.Add(y.direction.Mul(y.length))
	AB := cylend.Sub(y.position)
	AO := r.origin.Sub(y.position)

	AB_dot_d := AB.Dot(r.direction)
	AB_dot_AO := AB.Dot(AO)
	AB_dot_AB := AB.Dot(AB)

	m := AB_dot_d / AB_dot_AB
	n := AB_dot_AO / AB_dot_AB

	Q := r.direction.Sub(AB.Mul(m))
	R := AO.Sub(AB.Mul(n))

	a := Q.Dot(Q)
	b := 2.0 * Q.Dot(R)
	c := R.Dot(R) - y.radius*y.radius

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
		t = t0
	}

	if t > r.interDist {
		return false
	}

	t_k := t*m + n
	if t_k < 0.0 {
		// On start cap
		return false // y.intersectCap(r, i, true)
	}

	if t_k > 1.0 {
		// On end cap
		return false // y.intersectCap(r, i, false)
	}

	r.interDist = t
	r.interObj = i
	return true

}
func (y Cylinder) intersectCap(r *Ray, i int, start bool) bool {
	var pos Vector
	var dir Vector
	if start {
		pos = y.position
		dir = y.direction.Mul(-1)
	} else {
		pos = y.position.Add(y.direction.Mul(y.length))
		dir = y.direction
	}

	end := Plane{y.material, pos, dir, y.radius}
	return end.Intersect(r, i)
}

func (y Cylinder) getNormal(point Vector) Vector {
	PQ := point.Sub(y.position)
	pqa := PQ.Dot(y.direction)
	PQAA := y.direction.Mul(pqa)
	return PQ.Sub(PQAA).Normalize()
}
