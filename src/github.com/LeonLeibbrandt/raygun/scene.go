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
	CalcShadow   bool
	StartLine    int `json:"-"`
	EndLine      int `json:"-"`
	GridWidth    int `json:"-"`
	GridHeight   int `json:"-"`
	CameraPos    *Vector
	CameraLook   *Vector
	CameraUp     *Vector
	Look         *Vector `json:"-"`
	Vhor         *Vector `json:"-"`
	Vver         *Vector `json:"-"`
	Vp           *Vector `json:"-"`
	Image        *image.RGBA `json:"-"`
	GroupList    []*Group
	LightList    []Light
	MaterialList []*Material
	ImageList    map[string]image.Image `json:"-"`
}

func NewScene() *Scene {
	scn := &Scene{}

	scn.GroupList = make([]*Group, 0)
	scn.LightList = make([]Light, 0)
	scn.MaterialList = make([]*Material, 0)
	scn.ImageList = make(map[string]image.Image, 0)

	// defaults
	scn.ImgWidth = 320
	scn.ImgHeight = 200

	scn.TraceDepth = 3   // bounces
	scn.OverSampling = 1 // no OverSampling
	scn.VisionField = 60
	scn.CalcShadow = true
	return scn
}

func NewSceneFromFile(sceneFilename string) *Scene {
	scn := &Scene{}

	scn.GroupList = make([]*Group, 0)
	scn.LightList = make([]Light, 0)
	scn.MaterialList = make([]*Material, 0)
	scn.ImageList = make(map[string]image.Image, 0)

	// defaults
	scn.ImgWidth = 320
	scn.ImgHeight = 200

	scn.TraceDepth = 3   // bounces
	scn.OverSampling = 1 // no OverSampling
	scn.VisionField = 60
	scn.CalcShadow = true

	//scn.ObjectList = append(scn.ObjectList, Sphere{0,0.0,0.0,0.0,0.0})

	f, err := os.Open(sceneFilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 4*1024)

	scn.parseStream(r)

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

	scn.GroupList = make([]*Group, 0)
	scn.LightList = make([]Light, 0)
	scn.MaterialList = make([]*Material, 0)
	scn.ImageList = make(map[string]image.Image, 0)

	scn.Init()

	return scn
}

func NewSceneFromText(text string) *Scene {
	scn := &Scene{}
	scn.GroupList = make([]*Group, 0)
	scn.LightList = make([]Light, 0)
	scn.MaterialList = make([]*Material, 0)
	scn.ImageList = make(map[string]image.Image, 0)
	// defaults
	scn.GroupList = make([]*Group, 0)
	scn.ImgWidth = 320
	scn.ImgHeight = 200

	scn.TraceDepth = 3   // bounces
	scn.OverSampling = 1 // no OverSampling
	scn.VisionField = 60
	scn.CalcShadow = true

	r := bufio.NewReader(strings.NewReader(text))
	scn.parseStream(r)
	scn.Init()
	scn.CalcBounds()
	return scn
}

func (scn *Scene) AddGroup(group string) *Group {
	r := bufio.NewReader(strings.NewReader(group))
	scn.parseStream(r)
	return scn.GroupList[len(scn.GroupList)-1]
}

func (scn *Scene) parseStream(r *bufio.Reader) {
	groupIndex := len(scn.GroupList) - 1
	line, isPrefix, err := r.ReadLine()

	newplane := func(data []string) *Plane {
		mat, _ := strconv.Atoi(data[0])
		pos := ParseVector(data[1:4])
		nor := ParseVector(data[4:7])
		rad, _ := strconv.ParseFloat(data[7], 64)
		wid, _ := strconv.ParseFloat(data[8], 64)
		hei, _ := strconv.ParseFloat(data[9], 64)
		return NewPlane(pos.X, pos.Y, pos.Z, nor.X, nor.Y, nor.Z, rad, wid, hei, mat, scn)
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

		case "shadow":
			if data[0] == "true" {
				scn.CalcShadow = true
			} else {
				scn.CalcShadow = false
			}

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
			grp := NewGroup(data[0], pos.X, pos.Y, pos.Z, always, scn)
			grp.Bounds = plane
			scn.GroupList = append(scn.GroupList, grp)
			groupIndex = len(scn.GroupList) - 1

		case "sphere":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			rad, _ := strconv.ParseFloat(data[4], 64)

			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewSphere(pos.X, pos.Y, pos.Z, rad, mat, scn))

		case "plane":
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				newplane(data))

		case "texture":
			pos := ParseVector(data[0:3])
			nor := ParseVector(data[3:6])
			up := ParseVector(data[6:9])
			wid, _ := strconv.ParseFloat(data[9], 64)
			hei, _ := strconv.ParseFloat(data[10], 64)
			fn := data[11]
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewTexture(pos.X, pos.Y, pos.Z, nor.X, nor.Y, nor.Z, up.X, up.Y, up.Z, wid, hei, fn, scn))

		case "cube":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			width, _ := strconv.ParseFloat(data[4], 64)
			height, _ := strconv.ParseFloat(data[5], 64)
			depth, _ := strconv.ParseFloat(data[6], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewCube(pos.X, pos.Y, pos.Z, width, height, depth, mat, scn))

		case "cylinder":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			dir := ParseVector(data[4:7])
			len, _ := strconv.ParseFloat(data[7], 64)
			rad, _ := strconv.ParseFloat(data[8], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewCylinder(pos.X, pos.Y, pos.Z, dir.X, dir.Y, dir.Z, len, rad, mat, scn))

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
}

func (scn *Scene) Init() {

	scn.StartLine = 0 // Start rendering line
	scn.EndLine = scn.ImgHeight - 1

	scn.Image = image.NewRGBA(image.Rect(0, 0, scn.ImgWidth, scn.ImgHeight))

	scn.GridWidth = scn.ImgWidth * scn.OverSampling
	scn.GridHeight = scn.ImgHeight * scn.OverSampling

	scn.Look = scn.CameraLook.Sub(scn.CameraPos)
	scn.Vhor = scn.Look.Cross(scn.CameraUp)
	scn.Vhor = scn.Vhor.Normalize()

	scn.Vver = scn.Look.Cross(scn.Vhor)
	scn.Vver = scn.Vver.Normalize()

	fl := float64(scn.GridWidth) / (2 * math.Tan((0.5*scn.VisionField)*PI_180))

	Vp := scn.Look.Normalize()

	Vp.X = Vp.X*fl - 0.5*(float64(scn.GridWidth)*scn.Vhor.X+float64(scn.GridHeight)*scn.Vver.X)
	Vp.Y = Vp.Y*fl - 0.5*(float64(scn.GridWidth)*scn.Vhor.Y+float64(scn.GridHeight)*scn.Vver.Y)
	Vp.Z = Vp.Z*fl - 0.5*(float64(scn.GridWidth)*scn.Vhor.Z+float64(scn.GridHeight)*scn.Vver.Z)

	scn.Vp = Vp

}

func (scn *Scene) CalcBounds() {
	for _, grp := range scn.GroupList {
		grp.CalcBounds()
	}
}

func (scn *Scene) ObjectCount() int {
	count := 0
	for _, grp := range scn.GroupList {
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

func ParseMaterial(line []string) *Material {
	var f [6]float64
	for i, item := range line[3:8] {
		f[i], _ = strconv.ParseFloat(item, 64)
	}
	m, _ := NewMaterial(ParseColor(line[0:3]), f[0], f[1], f[2], f[3], f[4], f[5])
	return m
}
