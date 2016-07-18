package raygun

// RAY
type Ray struct {
	origin     *Vector
	direction  *Vector
	interDist  float64 // MAX_DIST
	interGrp   int
	interObj   int
	interColor Color
	a          float64
}

func NewRay(origin, direction *Vector) *Ray {
	r := &Ray{
		origin:    origin,
		direction: direction,
		interDist: MAX_DIST,
		interGrp:  -1,
		interObj:  -1,
	}
	r.a = r.direction.Dot(r.direction)
	return r
}
