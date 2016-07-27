package raygun

import (
)

// Material defines a raytracing material.
type Material struct {
	Color                                                              Color
	DifuseCol, SpecularCol, SpecularD, ReflectionCol, TransmitCol, IOR float64
}

func NewMaterial(color Color, difusecol, specularcol, speculard, reflectioncol, transmitcol, ior float64) (*Material, error) {
	m := &Material{
		Color:         color,
		DifuseCol:     difusecol,
		SpecularCol:   specularcol,
		SpecularD:     speculard,
		ReflectionCol: reflectioncol,
		TransmitCol:   transmitcol,
		IOR:           ior,

	}
	return m, nil

}

