package raygun

import (
	"image/color"
	"math"
)

// Color is an struct for managing colors.
type Color struct {
	R, G, B float64
}

// Add adds two colors together.
func (c Color) Add(u Color) Color {
	return Color{c.R + u.R, c.G + u.G, c.B + u.B}
}

// Mul multplies color components with a scalar.
func (c Color) Mul(f float64) Color {
	return Color{c.R * f, c.G * f, c.B * f}
}

// ToPixel return the standard color from a Color struct.
func (c Color) ToPixel() color.RGBA {
	c.R = math.Max(0.0, math.Min(c.R*255.0, 255.0))
	c.G = math.Max(0.0, math.Min(c.G*255.0, 255.0))
	c.B = math.Max(0.0, math.Min(c.B*255.0, 255.0))
	return color.RGBA{uint8(c.R), uint8(c.G), uint8(c.B), 255}
}
