package raygun

import (
	"math"
)

// Group is a single object in the scene composed of a group of primitives. It also acts as a first
// rejection mechanism in that the primitives are bounded by a sphere, or plane that is checked first
// for intersection.
type Group struct {
	Scene      *Scene `json:"-"`
	Name       string
	Center     *Vector
	ObjectList []Object
	Always     bool
	Bounds     GroupBounds `json:"-"`
}

// NewGroup creates a new group
func NewGroup(name string, x, y, z float64, always bool, scn *Scene) *Group {
	s := &Group{
		Scene:  scn,
		Name:   name,
		Center: &Vector{x, y, z},
		Always: always,
		//		ObjectList: ,
		Bounds: nil,
	}
	return s
}

// CalcBounds calculates the bounds of the group from the primitives if it is not always checked
// or the bounds object already exists.
func (g *Group) CalcBounds() {
	if g.Always {
		return
	}
	if g.Bounds != nil {
		return
	}
	max := 0.0
	for _, obj := range g.ObjectList {
		max = math.Max(max, obj.GetFurthest(g.Center))
	}
	if max == 0.0 {
		g.Always = true
		return
	}
	g.Bounds = NewSphere(g.Center.X, g.Center.Y, g.Center.Z, max, 0, g.Scene)
}

// HitBounds checks for intersection
func (g *Group) HitBounds(r *Ray) bool {
	if g.Always {
		return true
	}
	if g.Bounds == nil {
		return true
	}
	return g.Bounds.HitBounds(r)
}

// SetMaterial sets all the child primitives' material to the supplied value.
// This is an index into an array as defoined in scene.
func (g *Group) SetMaterial(m int) {
	for _, obj := range g.ObjectList {
		obj.SetMaterial(m)
	}
}
