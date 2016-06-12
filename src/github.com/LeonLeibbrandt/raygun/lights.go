package raygun

// Light defines a raytracing light at a specific position. Kind can be either point or ambient.
type Light struct {
	Position *Vector
	Color    Color
	Kind     string
}
