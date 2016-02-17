package raygun

import (
	"bufio"
	"fmt"
	"math"
)

type Group struct {
	Name       string
	Center     *Vector
	ObjectList []Object
	Always     bool
	Bounds     GroupBounds
}

func NewGroup(name string, x, y, z float64, always bool, objects []Object) *Group {
	s := &Group{
		Name:       name,
		Center:     &Vector{x, y, z},
		Always:     always,
		ObjectList: objects,
		Bounds:     nil,
	}
	return s
}

func (g *Group) CalcBounds() {
	if g.Always {
		return
	}
	if g.Bounds != nil {
		return
	}
	max := 0.0
	for _, obj := range g.ObjectList {
		max = math.Max(max, obj.Furthest(g.Center))
	}
	if max == 0.0 {
		g.Always = true
		return
	}
	g.Bounds = NewSphere(g.Center.x, g.Center.y, g.Center.z, max, 0)
}

func (g *Group) HitBounds(r *Ray) bool {
	if g.Always {
		return true
	}
	return g.Bounds.HitBounds(r)
}

func (g *Group) Write(buffer *bufio.Writer) {
	buffer.WriteString(fmt.Sprintf("%v %v\n", g.Name, g.Always))
}
