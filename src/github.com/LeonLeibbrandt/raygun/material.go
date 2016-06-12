package raygun


// Material defines a raytracing material.
type Material struct {
	Color                                                              Color
	DifuseCol, SpecularCol, SpecularD, ReflectionCol, TransmitCol, IOR float64
}
