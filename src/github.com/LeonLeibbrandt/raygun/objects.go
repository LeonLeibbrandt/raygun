package raygun

import (
	"image"
	"image/png"
	"math"
	"os"
	"strings"
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
	GetType() string
	GetMaterial() *Material
	SetMaterial(int)
	GetIntersect(r *Ray, g, i int) bool
	GetNormal(point *Vector) *Vector
	GetFurthest(point *Vector) float64
}

// Base has the default fields and method that is common to all primitives.
type Base struct {
	Scene         *Scene `json:"-"`
	Type          string
	MaterialIndex int
	Material      *Material `json:"-"`
}

func NewBase(scn *Scene, objtype string, m int) Base {
	return Base{
		Scene:         scn,
		Type:          objtype,
		MaterialIndex: m,
		Material:      scn.MaterialList[m],
	}
}

// Type defines the type of the object, it is used when parsing a scene file.
func (b *Base) GetType() string {
	return b.Type
}

// SetMaterial set the material index to the supplied value.
func (b *Base) SetMaterial(i int) {
	b.MaterialIndex = i
	b.Material = b.Scene.MaterialList[i]
}

func (b *Base) GetMaterial() *Material {
	return b.Material
}

// Sphere is a sphere with position and radius
type Sphere struct {
	Base
	Position *Vector
	Radius   float64
}

// NewSphere creates a new sphere from the supplied values.
func NewSphere(x, y, z, r float64, m int, scn *Scene) *Sphere {
	return &Sphere{
		Base:     NewBase(scn, "sphere", m),
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
func (e *Sphere) GetIntersect(r *Ray, g, i int) bool {
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
	r.interColor = e.Material.Color
	return true
}

func (e *Sphere) GetNormal(point *Vector) *Vector {
	normal := point.Sub(e.Position)
	return normal.Normalize()
}

// Furthest calculates the firhest distance this sphere can be from a the point.
// This is used to calculate the group bounds if this sphere is a child of a group.
func (e *Sphere) GetFurthest(point *Vector) float64 {
	return e.Position.Sub(point).Module() + e.Radius
}

// Plane is a primitive that has a position and a normal.
// It has been extended to be a disc - if Radius is set
// The planes horizontal is defined by the normal crossed with the scene Camera up, and the vertical
// as the plane normal crossed with horizontal. This is then used as the horisontal and vertical length:
// Horisontal = width
// Vertical = Height
// If Radius, Width and Height are zero then it is an infinite plane.

type Plane struct {
	Base
	Position   *Vector
	Normal     *Vector
	Horiz      *Vector `json:"-"`
	Vert       *Vector `json:"-"`
	Radius     float64
	Width      float64 // Only two of these will be used for plane, the one in the "normal" direction MUST be 0.0
	Height     float64
	halfWidth  float64 `json:"-"`
	halfHeight float64 `json:"-"`
}

func NewPlane(xp, yp, zp, xn, yn, zn, r, w, h float64, m int, scn *Scene) *Plane {
	p := &Plane{
		Base:       NewBase(scn, "plane", m),
		Position:   &Vector{xp, yp, zp},
		Normal:     (&Vector{xn, yn, zn}).Normalize(),
		Radius:     r,
		Width:      w,
		Height:     h,
		halfWidth:  (w / 2.0),
		halfHeight: (h / 2.0),
	}

	p.Horiz = p.Normal.Cross(scn.CameraUp).Normalize()
	if p.Horiz.Module() == 0.0 {
		p.Horiz = p.Normal.Cross(&Vector{0.0, -1.0, 0.0}).Normalize()
	}

	p.Vert = p.Normal.Cross(p.Horiz).Normalize()

	return p
}

func (p *Plane) HitBounds(r *Ray) bool {
	v := p.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}

	t := p.Normal.Dot(p.Position.Sub(r.origin)) / v
	if t < 0.0 || t > r.interDist {
		return false
	}

	interPoint := r.origin.Add(r.direction.Mul(t))

	u := interPoint.Sub(p.Position)

	if p.Radius > 0.0 {
		// We have a disc
		if u.Module() > p.Radius {
			return false
		}
	} else {
		if math.Abs(u.Dot(p.Horiz)) > p.halfWidth {
			return false
		}

		if math.Abs(u.Dot(p.Vert)) > p.halfHeight {
			return false
		}
	}
	return true
}

func (p *Plane) GetIntersect(r *Ray, g, i int) bool {
	v := p.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}

	t := p.Normal.Dot(p.Position.Sub(r.origin)) / v
	if t < 0.0 || t > r.interDist {
		return false
	}

	interPoint := r.origin.Add(r.direction.Mul(t))

	u := interPoint.Sub(p.Position)

	if p.Radius > 0.0 {
		// We have a disc
		if u.Module() > p.Radius {
			return false
		}
	} else {
		horiz := u.Dot(p.Horiz)
		if math.Abs(horiz) > p.halfWidth {
			return false
		}
		vert := u.Dot(p.Vert)
		if math.Abs(vert) > p.halfHeight {
			return false
		}
	}
	r.interColor = p.Material.Color
	r.interDist = t
	r.interObj = i
	r.interGrp = g
	return true

}

func (p *Plane) GetNormal(point *Vector) *Vector {
	return p.Normal
}

func (p *Plane) GetFurthest(point *Vector) float64 {
	dist := p.Position.Sub(point).Module()
	if p.Radius > 0.0 {
		return dist + p.Radius
	}
	return dist + math.Sqrt(p.halfWidth*p.halfWidth+p.halfHeight*p.halfHeight)
}

// Texture
type Texture struct {
	Scene      *Scene `json:"-"`
	Type       string
	Material   *Material `json:"-"`
	Position   *Vector
	Normal     *Vector
	Horiz      *Vector `json:"-"`
	Vert       *Vector `json:"-"`
	Width      float64 // Only two of these will be used for plane, the one in the "normal" direction MUST be 0.0
	Height     float64
	halfWidth  float64 `json:"-"`
	halfHeight float64 `json:"-"`
	ImageName  string
	Image      image.Image `json:"-"`
}

func NewTexture(xp, yp, zp, xn, yn, zn, w, h float64, filename string, scn *Scene) *Texture {
	mat, _ := NewMaterial(Color{1.0, 1.0, 1.0}, 1.0, 0.0, 0.0, 0.0, 0.0, 0.0)
	t := &Texture{
		Scene:      scn,
		Type:       "texture",
		Material:   mat,
		Position:   &Vector{xp, yp, zp},
		Normal:     (&Vector{xn, yn, zn}).Normalize(),
		Width:      w,
		Height:     h,
		halfWidth:  (w / 2.0),
		halfHeight: (h / 2.0),
		ImageName:  filename,
		Image:      nil,
	}
	if filename != "" {
		if strings.Contains(filename, ".png") {
			f, err := os.Open(filename)
			if err != nil {
				return nil
			}
			defer f.Close()
			t.Image, err = png.Decode(f)
			if err != nil {
				return nil
			}
		} else {
			// if not a file then an index into the scenes image list
			t.Image = scn.ImageList[filename]
		}
	}

	switch {
	case t.Normal.Eq(&Vector{1.0, 0.0, 0.0}):
		t.Horiz = &Vector{0.0, 1.0, 0.0}
		t.Vert = &Vector{0.0, 0.0, -1.0}
	case t.Normal.Eq(&Vector{0.0, 1.0, 0.0}):
		t.Horiz = &Vector{-1.0, 0.0, 0.0}
		t.Vert = &Vector{0.0, 0.0, -1.0}
	case t.Normal.Eq(&Vector{0.0, 0.0, 1.0}):
		t.Horiz = &Vector{-1.0, 0.0, 0.0}
		t.Vert = &Vector{0.0, 1.0, 0.0}
	default:
		vert := &Vector{0.0, 1.0, 0.0}
		t.Horiz = vert.Cross(t.Normal).Normalize()
		t.Vert = t.Horiz.Cross(t.Normal).Normalize()
	}
	/*
		t.Horiz = t.Normal.Cross(scn.CameraUp).Normalize()
		if t.Horiz.Module() == 0.0 {
			t.Horiz = t.Normal.Cross(&Vector{0.0, -1.0, 0.0}).Normalize()
		}
		t.Vert = t.Normal.Cross(t.Horiz).Normalize()
	*/
	return t
}

func (t *Texture) GetType() string {
	return t.Type
}

func (t *Texture) GetMaterial() *Material {
	return t.Material
}

func (t *Texture) SetMaterial(i int) {
}

func (t *Texture) HitBounds(r *Ray) bool {
	v := t.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}

	w := t.Normal.Dot(t.Position.Sub(r.origin)) / v
	if w < 0.0 || w > r.interDist {
		return false
	}

	interPoint := r.origin.Add(r.direction.Mul(w))

	u := interPoint.Sub(t.Position)

	if math.Abs(u.Dot(t.Horiz)) > t.halfWidth {
		return false
	}

	if math.Abs(u.Dot(t.Vert)) > t.halfHeight {
		return false
	}
	return true
}

func (t *Texture) GetIntersect(r *Ray, g, i int) bool {
	v := t.Normal.Dot(r.direction)
	if v == 0 {
		return false
	}

	w := t.Normal.Dot(t.Position.Sub(r.origin)) / v
	if w < 0.0 || w > r.interDist {
		return false
	}

	interPoint := r.origin.Add(r.direction.Mul(w))

	u := interPoint.Sub(t.Position)

	horiz := u.Dot(t.Horiz)
	if math.Abs(horiz) > t.halfWidth {
		return false
	}
	imgx := t.halfWidth + horiz

	vert := u.Dot(t.Vert)
	if math.Abs(vert) > t.halfHeight {
		return false
	}
	imgy := t.halfHeight + vert
	color, hasColor := t.getColor((t.Width-imgx)/t.Width, imgy/t.Height)
	if !hasColor {
		return false
	}

	r.interColor = color
	r.interDist = w
	r.interObj = i
	r.interGrp = g
	return true

}

func (t *Texture) GetNormal(point *Vector) *Vector {
	return t.Normal
}

func (t *Texture) GetFurthest(point *Vector) float64 {
	dist := t.Position.Sub(point).Module()
	return dist + math.Sqrt(t.halfWidth*t.halfWidth+t.halfHeight*t.halfHeight)
}

func (t *Texture) getColor(x, y float64) (Color, bool) {
	bounds := t.Image.Bounds()
	imgx := x*float64(bounds.Max.X-bounds.Min.X) + float64(bounds.Min.X)
	imgy := y*float64(bounds.Max.Y-bounds.Min.Y) + float64(bounds.Min.Y)
	col := t.Image.At(int(imgx), int(imgy))
	r, g, b, a := col.RGBA()
	fr := float64(r)
	fg := float64(g)
	fb := float64(b)
	fa := float64(a)
	if a == 0x0000 {
		return Color{fr / fa, fg / fa, fb / fa}, false
	}
	return Color{fr / fa, fg / fa, fb / fa}, true // FromColor(mat.Image.At(int(imgx), int(imgy)))
}

// Cube

type Cube struct {
	Base
	Position *Vector
	Width    float64
	Height   float64
	Depth    float64
	Min      *Vector `json:"-"`
	Max      *Vector `json:"-"`
}

func NewCube(x, y, z, w, h, d float64, m int, scn *Scene) *Cube {
	c := &Cube{
		Base:     NewBase(scn, "cube", m),
		Position: &Vector{x, y, z},
		Width:    w, // x direction
		Height:   h, // y direction
		Depth:    d, // z direction
	}
	c.initMinMax()
	return c
}

func (c *Cube) GetIntersect(r *Ray, g, i int) bool {
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
		r.interColor = c.Material.Color
		return true
	}
	return false
}

func (c *Cube) GetNormal(point *Vector) *Vector {

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

func (c *Cube) GetFurthest(point *Vector) float64 {
	max := math.Max(math.Max(c.Width/2.0, c.Height/2.), c.Depth)
	return c.Position.Sub(point).Module() + 1.5*max
}

type Cylinder struct {
	Base
	Position  *Vector
	Direction *Vector
	Length    float64
	Radius    float64
	StartDisc *Plane `json:"-"`
	EndDisc   *Plane `json:"-"`
}

func NewCylinder(xp, yp, zp, xd, yd, zd, l, r float64, m int, scn *Scene) *Cylinder {
	c := &Cylinder{
		Base:      NewBase(scn, "cylinder", m),
		Position:  &Vector{xp, yp, zp},
		Direction: (&Vector{xd, yd, zd}).Normalize(),
		Length:    l,
		Radius:    r,
	}
	pos := c.Position
	dir := c.Direction.Mul(-1)
	c.StartDisc = NewPlane(pos.X, pos.Y, pos.Z, dir.X, dir.Y, dir.Z, c.Radius, 0.0, 0.0, c.MaterialIndex, c.Scene)

	pos = c.Position.Add(c.Direction.Mul(c.Length))
	dir = c.Direction
	c.EndDisc = NewPlane(pos.X, pos.Y, pos.Z, dir.X, dir.Y, dir.Z, c.Radius, 0.0, 0.0, c.MaterialIndex, c.Scene)
	return c
}

// http://blog.makingartstudios.com/?p=286
func (y *Cylinder) GetIntersect(r *Ray, g, i int) bool {
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
	r.interColor = y.Material.Color
	return true

}
func (y *Cylinder) intersectCap(r *Ray, start bool) bool {
	if start {
		return y.StartDisc.GetIntersect(r, 0, 0)
	}
	return y.EndDisc.GetIntersect(r, 0, 0)
}

func (y *Cylinder) GetNormal(point *Vector) *Vector {
	PQ := point.Sub(y.Position)
	pqa := PQ.Dot(y.Direction)
	PQAA := y.Direction.Mul(pqa)
	return PQ.Sub(PQAA).Normalize()
}

func (y *Cylinder) GetFurthest(point *Vector) float64 {
	return y.Position.Sub(point).Module() + y.Length + y.Radius
}
