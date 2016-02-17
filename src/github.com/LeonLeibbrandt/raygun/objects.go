package raygun

import (
	"bufio"
	"fmt"
	"math"
)

const EPS = 1e-9

// http://www.hugi.scene.org/online/coding/hugi%2024%20-%20coding%20graphics%20chris%20dragan%20raytracing%20shapes.htm

type GroupBounds interface {
	HitBounds(r *Ray) bool
}

// Object
type Object interface {
	Type() string
	Material() int
	SetMaterial(int)
	Intersect(r *Ray, g, i int) bool
	getNormal(point *Vector) *Vector
	Furthest(point *Vector) float64
	Write(*bufio.Writer)
}

type Base struct {
	objecttype string
	material   int
}

func (b *Base) Type() string {
	return b.objecttype
}

func (b *Base) Material() int {
	return b.material
}

func (b *Base) SetMaterial(i int) {
	b.material = i
}

// Sphere
type Sphere struct {
	Base
	Position *Vector
	Radius   float64
}

func NewSphere(x, y, z, r float64, m int) *Sphere {
	return &Sphere{
		Base: Base{
			objecttype: "sphere",
			material:   m,
		},
		Position: &Vector{x, y, z},
		Radius:   r,
	}
}

func (e *Sphere) HitBounds(r *Ray) bool {
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
	return true
}

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

func (e *Sphere) Furthest(point *Vector) float64 {
	return e.Position.Sub(point).Module() + e.Radius
}

func (e *Sphere) String() string {
	return fmt.Sprintf("<Esf: %d %s %.2f>", e.Material, e.Position.String(), e.Radius)
}

func (e *Sphere) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewSphere(%.2f, %.2f, %.2f, %.2f, %v),\n",
		e.Position.x,
		e.Position.y,
		e.Position.z,
		e.Radius,
		e.Material()))
}

// PLANE
type Plane struct {
	Base
	Position   *Vector
	Normal     *Vector
	Radius     float64
	Width      float64 // Only two of these will be used for plane, the one in the "normal" direction MUST be 0.0
	Height     float64
	Depth      float64
	halfWidth  float64
	halfHeight float64
	halfDepth  float64
}

func NewPlane(xp, yp, zp, xn, yn, zn, r, w, h, d float64, m int) *Plane {
	p := &Plane{
		Base: Base{
			objecttype: "plane",
			material:   m,
		},
		Position:   &Vector{xp, yp, zp},
		Normal:     (&Vector{xn, yn, zn}).Normalize(),
		Radius:     r,
		Width:      w,
		Height:     h,
		Depth:      d,
		halfWidth:  w / 2.0,
		halfHeight: h / 2.0,
		halfDepth:  d / 2.0,
	}
	// fmt.Printf("%#v\n", p)
	return p
}

var done = false

func (p *Plane) HitBounds(r *Ray) bool {
	v := p.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}
	t := p.Normal.Dot(p.Position.Sub(r.origin)) / v
	if t < 0.0 {
		return false
	}
	// We have a finite plane
	interPoint := r.origin.Add(r.direction.Mul(t))
	if p.halfWidth > 0.0 && (interPoint.x < p.Position.x-p.halfWidth || interPoint.x > p.Position.x+p.halfWidth) {
		return false
	}

	if p.halfHeight > 0.0 && (interPoint.y < p.Position.y-p.halfHeight || interPoint.y > p.Position.y+p.halfHeight) {
		return false
	}

	if p.halfDepth > 0.0 && (interPoint.z < p.Position.z-p.halfDepth || interPoint.z > p.Position.z+p.halfDepth) {
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
		if interPoint.x < p.Position.x-p.halfWidth || interPoint.x > p.Position.x+p.halfWidth ||
			interPoint.y < p.Position.y-p.halfHeight || interPoint.y > p.Position.y+p.halfHeight ||
			interPoint.z < p.Position.z-p.halfDepth || interPoint.z > p.Position.z+p.halfDepth {
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

func (p *Plane) String() string {
	return fmt.Sprintf("<Pla: %d %s %.2f>", p.Material, p.Normal.String(), p.Radius)
}

func (p *Plane) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewPlane(%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %v),\n",
		p.Position.x,
		p.Position.y,
		p.Position.z,
		p.Normal.x,
		p.Normal.y,
		p.Normal.z,
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
	min      *Vector
	max      *Vector
}

func NewCube(x, y, z, w, h, d float64, m int) *Cube {
	c := &Cube{
		Base: Base{
			objecttype: "cube",
			material:   m,
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
		r.interGrp = g
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

func (c *Cube) Furthest(point *Vector) float64 {
	max := math.Max(math.Max(c.Width/2.0, c.Height/2.), c.Depth)
	return c.Position.Sub(point).Module() + max
}

func (c *Cube) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("raygun.NewCube(%.2f, %.2f, %.2f, %.2f, %.2f, %.2f, %v),\n",
		c.Position.x,
		c.Position.y,
		c.Position.z,
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
	startDisc *Plane
	endDisc   *Plane
}

func NewCylinder(xp, yp, zp, xd, yd, zd, l, r float64, m int) *Cylinder {
	c := &Cylinder{
		Base: Base{
			objecttype: "cylinder",
			material:   m,
		},
		Position:  &Vector{xp, yp, zp},
		Direction: (&Vector{xd, yd, zd}).Normalize(),
		Length:    l,
		Radius:    r,
	}
	pos := c.Position
	dir := c.Direction.Mul(-1)
	c.startDisc = NewPlane(pos.x, pos.y, pos.z, dir.x, dir.y, dir.z, c.Radius, 0.0, 0.0, 0.0, c.Material())

	pos = c.Position.Add(c.Direction.Mul(c.Length))
	dir = c.Direction
	c.endDisc = NewPlane(pos.x, pos.y, pos.z, dir.x, dir.y, dir.z, c.Radius, 0.0, 0.0, 0.0, c.Material())
	return c
}

// http://blog.makingartstudios.com/?p=286
func (y *Cylinder) Intersect(r *Ray, g, i int) bool {
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
		if y.intersectCap(r, true) && t1 > r.interDist {
			t = t1
		} else {
			return false // y.intersectCap(r, i, true)
		}
	}

	if t_k > 1.0 {
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
		return y.startDisc.Intersect(r, 0, 0)
	}
	return y.endDisc.Intersect(r, 0, 0)
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
		y.Position.x,
		y.Position.y,
		y.Position.z,
		y.Direction.x,
		y.Direction.y,
		y.Direction.z,
		y.Length,
		y.Radius,
		y.Material(),
	))
}
