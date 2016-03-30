package raygun

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// SCENE
type Scene struct {
	ImgWidth     int
	ImgHeight    int
	TraceDepth   int
	OverSampling int
	VisionField  float64
	StartLine    int
	EndLine      int
	GridWidth    int
	GridHeight   int
	CameraPos    *Vector
	CameraLook   *Vector
	CameraUp     *Vector
	Look         *Vector
	Vhor         *Vector
	Vver         *Vector
	Vp           *Vector
	Image        *image.RGBA
	GroupList    []*Group
	LightList    []Light
	MaterialList []Material
}

func NewScene(sceneFilename string) *Scene {
	scn := &Scene{}
	// defaults
	scn.GroupList = make([]*Group, 0)
	groupIndex := -1
	scn.ImgWidth = 320
	scn.ImgHeight = 200

	scn.TraceDepth = 3   // bounces
	scn.OverSampling = 1 // no OverSampling
	scn.VisionField = 60

	//scn.ObjectList = append(scn.ObjectList, Sphere{0,0.0,0.0,0.0,0.0})

	f, err := os.Open(sceneFilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 4*1024)
	if err != nil {
		panic(err)
	}
	line, isPrefix, err := r.ReadLine()

	newplane := func(data []string) *Plane {
		mat, _ := strconv.Atoi(data[0])
		pos := ParseVector(data[1:4])
		nor := ParseVector(data[4:7])
		rad, _ := strconv.ParseFloat(data[7], 64)
		wid, _ := strconv.ParseFloat(data[8], 64)
		hei, _ := strconv.ParseFloat(data[9], 64)
		dep, _ := strconv.ParseFloat(data[10], 64)
		return NewPlane(pos.X, pos.Y, pos.Z, nor.X, nor.Y, nor.Z, rad, wid, hei, dep, mat)
	}

	for err == nil && !isPrefix {

		s := string(line)
		if len(s) == 0 {
			line, isPrefix, err = r.ReadLine()
			continue
		}

		if s[0:1] == "#" {
			line, isPrefix, err = r.ReadLine()
			continue
		}

		sline := strings.Split(s, " ")
		word := sline[0]
		untrimmed := sline[1:]
		var data []string

		for _, item := range untrimmed {
			if item == "" || item == " " {
				continue
			}
			data = append(data, strings.Trim(item, " "))
		}

		switch word {
		case "size":
			scn.ImgWidth, _ = strconv.Atoi(data[0])
			scn.ImgHeight, _ = strconv.Atoi(data[1])
			scn.EndLine = scn.ImgHeight - 1 // End rendering line
		case "nbounces":
			scn.TraceDepth, _ = strconv.Atoi(data[0]) // n. bounces
		case "oversampling":
			scn.OverSampling, _ = strconv.Atoi(data[0])
		case "vision":
			scn.VisionField, _ = strconv.ParseFloat(data[0], 64)
		case "renderslice":
			scn.StartLine, _ = strconv.Atoi(data[0])
			scn.EndLine, _ = strconv.Atoi(data[1])

		case "cameraPos":
			scn.CameraPos = ParseVector(data)
		case "cameraLook":
			scn.CameraLook = ParseVector(data)
		case "cameraUp":
			scn.CameraUp = ParseVector(data)

		case "group":
			var plane GroupBounds
			plane = nil
			if len(data) == 16 {
				plane = newplane(data[5:])
			}
			pos := ParseVector(data[1:4])
			always := true
			if data[4] == "false" {
				always = false
			}
			grp := NewGroup(data[0], pos.X, pos.Y, pos.Z, always)
			grp.Bounds = plane
			scn.GroupList = append(scn.GroupList, grp)
			groupIndex = groupIndex + 1

		case "sphere":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			rad, _ := strconv.ParseFloat(data[4], 64)

			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewSphere(pos.X, pos.Y, pos.Z, rad, mat))

		case "plane":
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				newplane(data))

		case "cube":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			width, _ := strconv.ParseFloat(data[4], 64)
			height, _ := strconv.ParseFloat(data[5], 64)
			depth, _ := strconv.ParseFloat(data[6], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewCube(pos.X, pos.Y, pos.Z, width, height, depth, mat))

		case "cylinder":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			dir := ParseVector(data[4:7])
			len, _ := strconv.ParseFloat(data[7], 64)
			rad, _ := strconv.ParseFloat(data[8], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewCylinder(pos.X, pos.Y, pos.Z, dir.X, dir.Y, dir.Z, len, rad, mat))

		case "light":
			light := Light{ParseVector(data[0:3]), ParseColor(data[3:6]), data[6]}
			scn.LightList = append(scn.LightList, light)

		case "material":
			mat := ParseMaterial(data)
			scn.MaterialList = append(scn.MaterialList, mat)

		}
		line, isPrefix, err = r.ReadLine()
	}

	if isPrefix {
		panic(errors.New("buffer size to small"))
	}
	if err != io.EOF {
		panic(err)
	}

	scn.Init()

	scn.CalcBounds()

	return scn
}

func NewSceneFromParams(imgWidth, imgHeight, traceDepth, overSampling int,
	visionField float64,
	cameraPos, cameraLook, cameraUp *Vector) *Scene {
	scn := &Scene{
		ImgWidth:     imgWidth,
		ImgHeight:    imgHeight,
		TraceDepth:   traceDepth,
		OverSampling: overSampling,
		VisionField:  visionField,
		CameraPos:    cameraPos,
		CameraLook:   cameraLook,
		CameraUp:     cameraUp,
	}

	scn.Init()

	return scn
}

func (sc *Scene) Init() {

	sc.StartLine = 0 // Start rendering line
	sc.EndLine = sc.ImgHeight - 1

	sc.Image = image.NewRGBA(image.Rect(0, 0, sc.ImgWidth, sc.ImgHeight))

	sc.GridWidth = sc.ImgWidth * sc.OverSampling
	sc.GridHeight = sc.ImgHeight * sc.OverSampling

	sc.Look = sc.CameraLook.Sub(sc.CameraPos)
	sc.Vhor = sc.Look.Cross(sc.CameraUp)
	sc.Vhor = sc.Vhor.Normalize()

	sc.Vver = sc.Look.Cross(sc.Vhor)
	sc.Vver = sc.Vver.Normalize()

	fl := float64(sc.GridWidth) / (2 * math.Tan((0.5*sc.VisionField)*PI_180))

	Vp := sc.Look.Normalize()

	Vp.X = Vp.X*fl - 0.5*(float64(sc.GridWidth)*sc.Vhor.X+float64(sc.GridHeight)*sc.Vver.X)
	Vp.Y = Vp.Y*fl - 0.5*(float64(sc.GridWidth)*sc.Vhor.Y+float64(sc.GridHeight)*sc.Vver.Y)
	Vp.Z = Vp.Z*fl - 0.5*(float64(sc.GridWidth)*sc.Vhor.Z+float64(sc.GridHeight)*sc.Vver.Z)

	sc.Vp = Vp

}

func (sc *Scene) CalcBounds() {
	for _, grp := range sc.GroupList {
		grp.CalcBounds()
	}
}

func (sc *Scene) ObjectCount() int {
	count := 0
	for _, grp := range sc.GroupList {
		count = count + len(grp.ObjectList)
	}
	return count
}

func (sc *Scene) String() string {
	return fmt.Sprintf("var scene = &raygun.Scene{}\n")
}

// Auxiliary Methods
func ParseVector(line []string) *Vector {
	x, _ := strconv.ParseFloat(line[0], 64)
	y, _ := strconv.ParseFloat(line[1], 64)
	z, _ := strconv.ParseFloat(line[2], 64)
	return &Vector{x, y, z}
}

func ParseColor(line []string) Color {
	r, _ := strconv.ParseFloat(line[0], 64)
	g, _ := strconv.ParseFloat(line[1], 64)
	b, _ := strconv.ParseFloat(line[2], 64)
	return Color{r, g, b}
}

func ParseMaterial(line []string) Material {
	var f [6]float64
	for i, item := range line[3:] {
		f[i], _ = strconv.ParseFloat(item, 64)
	}
	return Material{ParseColor(line[0:3]), f[0], f[1], f[2], f[3], f[4], f[5]}
}
