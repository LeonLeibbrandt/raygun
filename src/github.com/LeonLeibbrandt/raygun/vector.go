package raygun

import (
	"fmt"
	"math"
)

type Vector struct {
	x, y, z float64
}

// v.Dot(u) -> float64
func (v *Vector) Dot(u *Vector) float64 {
	return (v.x*u.x + v.y*u.y + v.z*u.z)
}

func (v *Vector) Cross(u *Vector) (*Vector) {
	r := &Vector{}
	r.x = u.y*v.z - u.z*v.y
	r.y = u.z*v.z - u.x*v.z
	r.z = u.x*v.y - u.y*v.x
	return r
}

func (v *Vector) Module() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v *Vector) Normalize() *Vector {
	if m := v.Module(); m != 0.0 {
		return &Vector{v.x / m, v.y / m, v.z / m}
	}
	return v
}

func (v *Vector) Add(u *Vector) *Vector {
	return &Vector{v.x + u.x, v.y + u.y, v.z + u.z}
}

func (v *Vector) Sub(u *Vector) *Vector {
	return &Vector{v.x - u.x, v.y - u.y, v.z - u.z}
}

func (v *Vector) Mul(u float64) *Vector {
	return &Vector{v.x * u, v.y * u, v.z * u}
}

func (v *Vector) Div(u *Vector) *Vector {
	return &Vector{v.x / u.x, v.y / u.y, v.z / u.z}
}

func (v *Vector) Min(u *Vector) *Vector {
	return &Vector{math.Min(v.x, u.x), math.Min(v.y, u.y), math.Min(v.z, u.z)}
}

func (v *Vector) Max(u *Vector) *Vector {
	return &Vector{math.Max(v.x, u.x), math.Max(v.y, u.y), math.Max(v.z, u.z)}
}

func (v *Vector) String() string {
	return fmt.Sprintf("&Vector{ %.2f, %.2f, %.2f},", v.x, v.y, v.z)
}
