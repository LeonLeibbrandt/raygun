package raygun

import (
	"math"
)

type Vector struct {
	X, Y, Z float64
}

// v.Dot(u) -> float64
func (v *Vector) Dot(u *Vector) float64 {
	return (v.X*u.X + v.Y*u.Y + v.Z*u.Z)
}

func (v *Vector) Cross(u *Vector) (*Vector) {
	r := &Vector{}
	r.X = u.Y*v.Z - u.Z*v.Y
	r.Y = u.Z*v.X - u.X*v.Z
	r.Z = u.X*v.Y - u.Y*v.X
	return r
}

func (v *Vector) Module() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v *Vector) Normalize() *Vector {
	if m := v.Module(); m != 0.0 {
		return &Vector{v.X / m, v.Y / m, v.Z / m}
	}
	return v
}

func (v *Vector) Add(u *Vector) *Vector {
	return &Vector{v.X + u.X, v.Y + u.Y, v.Z + u.Z}
}

func (v *Vector) Sub(u *Vector) *Vector {
	return &Vector{v.X - u.X, v.Y - u.Y, v.Z - u.Z}
}

func (v *Vector) Mul(u float64) *Vector {
	return &Vector{v.X * u, v.Y * u, v.Z * u}
}

func (v *Vector) Div(u *Vector) *Vector {
	return &Vector{v.X / u.X, v.Y / u.Y, v.Z / u.Z}
}

func (v *Vector) Min(u *Vector) *Vector {
	return &Vector{math.Min(v.X, u.X), math.Min(v.Y, u.Y), math.Min(v.Z, u.Z)}
}

func (v *Vector) Max(u *Vector) *Vector {
	return &Vector{math.Max(v.X, u.X), math.Max(v.Y, u.Y), math.Max(v.Z, u.Z)}
}

// func (v *Vector) String() string {
// 	return fmt.Sprintf("&Vector{ %.2f, %.2f, %.2f},", v.X, v.Y, v.Z)
// }
