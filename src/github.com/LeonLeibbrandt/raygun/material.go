package raygun

import (
	"image"
	"image/png"
	"os"
)

// Material defines a raytracing material.
type Material struct {
	Color                                                              Color
	DifuseCol, SpecularCol, SpecularD, ReflectionCol, TransmitCol, IOR float64
	FileName                                                           string
	Image                                                              image.Image `json:"-"`
}

func NewMaterial(color Color, difusecol, specularcol, speculard, reflectioncol, transmitcol, ior float64, filename string) (*Material, error) {
	m := &Material{
		Color:         color,
		DifuseCol:     difusecol,
		SpecularCol:   specularcol,
		SpecularD:     speculard,
		ReflectionCol: reflectioncol,
		TransmitCol:   transmitcol,
		IOR:           ior,
		FileName:      filename,
		Image:         nil,
	}
	if filename != "" {
		f, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		m.Image, err = png.Decode(f)
		if err != nil {
			return nil, err
		}
	}
	return m, nil

}

/*
func (m *Material) GetColor(x, y int) (c Color) {
	if m.Image != nil {
		return FromColor(m.Image.At(x, y))
	}
	return m.Color
}
*/
