package raygun

import (
	"bufio"
	"errors"
	"image"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// SCENE
type Scene struct {
	imgWidth     int
	imgHeight    int
	traceDepth   int
	oversampling int
	visionField  float64
	startline    int
	endline      int
	gridWidth    int
	gridHeight   int
	cameraPos    *Vector
	cameraLook   *Vector
	cameraUp     *Vector
	look         *Vector
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
	scn.imgWidth = 320
	scn.imgHeight = 200

	scn.traceDepth = 3   // bounces
	scn.oversampling = 1 // no oversampling
	scn.visionField = 60

	scn.startline = 0 // Start rendering line
	scn.endline = scn.imgHeight - 1

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
			scn.imgWidth, _ = strconv.Atoi(data[0])
			scn.imgHeight, _ = strconv.Atoi(data[1])
			scn.endline = scn.imgHeight - 1 // End rendering line
		case "nbounces":
			scn.traceDepth, _ = strconv.Atoi(data[0]) // n. bounces
		case "oversampling":
			scn.oversampling, _ = strconv.Atoi(data[0])
		case "vision":
			scn.visionField, _ = strconv.ParseFloat(data[0], 64)
		case "renderslice":
			scn.startline, _ = strconv.Atoi(data[0])
			scn.endline, _ = strconv.Atoi(data[1])

		case "cameraPos":
			scn.cameraPos = ParseVector(data)
		case "cameraLook":
			scn.cameraLook = ParseVector(data)
		case "cameraUp":
			scn.cameraUp = ParseVector(data)

		case "group":
			pos := ParseVector(data[1:4])
			always := true
			if data[4] == "false" {
				always = false
			}
			grp := NewGroup(data[0], pos.x, pos.y, pos.z, always, []Object{})
			switch grp.Name {
			case "fence1":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 1.0, 0.0, 0.0, 0.0, 0.0, 12.2, 2.7, 0)
			case "fence2":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 1.0, 0.0, 0.0, 0.0, 0.0, 12.2, 2.7, 0)
			case "fence3":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 0.0, 1.0, 0.0, 0.0, 12.2, 0.0, 2.7, 0)
			case "fence4":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 0.0, 1.0, 0.0, 0.0, 8.2, 0.0, 2.8, 0)
			case "fence5":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 0.0, 1.0, 0.0, 0.0, 2.2, 0.0, 2.8, 0)
			case "closedgate":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 0.0, 1.0, 0.0, 0.0, 2.2, 0.0, 2.8, 0)
			case "opengate":
				grp.Bounds = NewPlane(pos.x, pos.y, pos.z, 1.0, 1.0, 0.0, 0.0, 2.2, 0.0, 4.5, 0)
			}
			scn.GroupList = append(scn.GroupList, grp)
			groupIndex = groupIndex + 1

		case "sphere":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			rad, _ := strconv.ParseFloat(data[4], 64)

			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewSphere(pos.x, pos.y, pos.z, rad, mat))

		case "plane":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			nor := ParseVector(data[4:7])
			rad, _ := strconv.ParseFloat(data[7], 64)
			wid, _ := strconv.ParseFloat(data[8], 64)
			hei, _ := strconv.ParseFloat(data[9], 64)
			dep, _ := strconv.ParseFloat(data[10], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewPlane(pos.x, pos.y, pos.z, nor.x, nor.y, nor.z, rad, wid, hei, dep, mat))

		case "cube":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			width, _ := strconv.ParseFloat(data[4], 64)
			height, _ := strconv.ParseFloat(data[5], 64)
			depth, _ := strconv.ParseFloat(data[6], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewCube(pos.x, pos.y, pos.z, width, height, depth, mat))

		case "cylinder":
			mat, _ := strconv.Atoi(data[0])
			pos := ParseVector(data[1:4])
			dir := ParseVector(data[4:7])
			len, _ := strconv.ParseFloat(data[7], 64)
			rad, _ := strconv.ParseFloat(data[8], 64)
			scn.GroupList[groupIndex].ObjectList = append(scn.GroupList[groupIndex].ObjectList,
				NewCylinder(pos.x, pos.y, pos.z, dir.x, dir.y, dir.z, len, rad, mat))

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

	scn.Image = image.NewRGBA(image.Rect(0, 0, scn.imgWidth, scn.imgHeight))

	scn.gridWidth = scn.imgWidth * scn.oversampling
	scn.gridHeight = scn.imgHeight * scn.oversampling

	scn.look = scn.cameraLook.Sub(scn.cameraPos)
	scn.Vhor = scn.look.Cross(scn.cameraUp)
	scn.Vhor = scn.Vhor.Normalize()

	scn.Vver = scn.look.Cross(scn.Vhor)
	scn.Vver = scn.Vver.Normalize()

	fl := float64(scn.gridWidth) / (2 * math.Tan((0.5*scn.visionField)*PI_180))

	Vp := scn.look.Normalize()

	Vp.x = Vp.x*fl - 0.5*(float64(scn.gridWidth)*scn.Vhor.x+float64(scn.gridHeight)*scn.Vver.x)
	Vp.y = Vp.y*fl - 0.5*(float64(scn.gridWidth)*scn.Vhor.y+float64(scn.gridHeight)*scn.Vver.y)
	Vp.z = Vp.z*fl - 0.5*(float64(scn.gridWidth)*scn.Vhor.z+float64(scn.gridHeight)*scn.Vver.z)

	scn.Vp = Vp

	for _, grp := range scn.GroupList {
		grp.CalcBounds()
	}
	return scn
}

func (sc *Scene) ObjectCount() int {
	count := 0
	for _, grp := range sc.GroupList {
		count = count + len(grp.ObjectList)
	}
	return count
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
