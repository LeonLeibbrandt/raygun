package raygun

import (
	"bytes"
	"fmt"
	// "image/png"
	"io/ioutil"
	"math"
	"os"
	"time"
)

const (
	MAX_DIST = 1999999999
	PI_180   = 0.017453292
	SMALL    = 0.000000001
)

type RayGun struct {
	FileName   string
	NumWorkers int
	Scene      *Scene
	Done       chan bool
	Line       chan int
}

func NewRayGun(filename string, numworkers int) (*RayGun, error) {
	rg := &RayGun{
		FileName:   filename,
		NumWorkers: numworkers,
		Scene:      NewSceneFromFile(filename),
		Done:       make(chan bool, numworkers),
		Line:       make(chan int),
	}
	return rg, nil
}

func NewRayGunFromScene(scene *Scene, numworkers int) (*RayGun, error) {
	rg := &RayGun{
		NumWorkers: numworkers,
		Scene:      scene,
		Done:       make(chan bool, numworkers),
		Line:       make(chan int),
	}
	return rg, nil
}

// Render renders the scene and returns a jpeg that has been base64 encoded
func (rg *RayGun) Render() {
	start := time.Now()
	for i := 0; i < rg.NumWorkers; i++ {
		go rg.renderPixel(rg.Line, rg.Done)
	}

	fmt.Printf("Rendering: %s\n", rg.FileName)
	fmt.Printf("Line (from %d to %d): ", rg.Scene.StartLine, rg.Scene.EndLine)

	for y := rg.Scene.StartLine; y < rg.Scene.EndLine; y++ {
		rg.Line <- y
	}
	close(rg.Line)

	// wait for all workers to finish
	for i := 0; i < rg.NumWorkers; i++ {
		<-rg.Done
	}
	/*
		output, err := os.Create(rg.FileName + ".png")
		if err != nil {
			panic(err)
		}

		err = png.Encode(output, rg.Scene.Image)
		if err != nil {
			panic(err)
		}
	*/
	elapsed := time.Since(start)
	fmt.Printf("\nTime %s for %v objects\n", elapsed, rg.Scene.ObjectCount())
}

func (rg *RayGun) calcShadow(r *Ray, collisionObj, collisionGrp int) float64 {
	shadow := 1.0 //starts with no shadow
	for g, grp := range rg.Scene.GroupList {
		for i, obj := range grp.ObjectList {
			r.interObj = -1
			r.interGrp = -1
			r.interDist = MAX_DIST

			grpCheck := true

			if len(rg.Scene.GroupList) > 1 && g == collisionGrp {
				grpCheck = false
			}

			if obj.GetIntersect(r, g, i) && grpCheck && i != collisionObj {
				shadow *= obj.GetMaterial().TransmitCol
			}
		}
	}
	return shadow
}

func (rg *RayGun) trace(r *Ray, depth int) (c Color) {
	for g, grp := range rg.Scene.GroupList {
		if !grp.HitBounds(r) {
			continue
		}
		for i, obj := range grp.ObjectList {
			obj.GetIntersect(r, g, i)
		}
	}

	if r.interObj >= 0 {
		obj := rg.Scene.GroupList[r.interGrp].ObjectList[r.interObj]
		material := obj.GetMaterial()
		interPoint := r.origin.Add(r.direction.Mul(r.interDist))
		incidentV := interPoint.Sub(r.origin)
		originBackV := r.direction.Mul(-1.0)
		originBackV = originBackV.Normalize()
		vNormal := rg.Scene.GroupList[r.interGrp].ObjectList[r.interObj].GetNormal(interPoint)
		for _, light := range rg.Scene.LightList {
			switch light.Kind {
			case "ambient":
				c = c.Add(light.Color)
			case "point":
				lightDir := light.Position.Sub(interPoint)
				lightDir = lightDir.Normalize()
				lightRay := NewRay(interPoint, lightDir)
				shadow := 1.0
				if rg.Scene.CalcShadow {
					shadow = rg.calcShadow(lightRay, r.interObj, r.interGrp)
				}
				NL := vNormal.Dot(lightDir)

				if NL > 0.0 {
					if material.DifuseCol > 0.0 { // ------- Difuso
						difuseColor := light.Color.Mul(material.DifuseCol).Mul(NL)
						difuseColor.R *= r.interColor.R * shadow
						difuseColor.G *= r.interColor.G * shadow
						difuseColor.B *= r.interColor.B * shadow
						c = c.Add(difuseColor)
					}
					if material.SpecularCol > 0.0 { // ----- Especular
						R := (vNormal.Mul(2).Mul(NL)).Sub(lightDir)
						spec := originBackV.Dot(R)
						if spec > 0.0 {
							spec = material.SpecularCol * math.Pow(spec, material.SpecularD)
							specularColor := light.Color.Mul(spec).Mul(shadow)
							c = c.Add(specularColor)
						}
					}
				}
			}
		}
		if depth < rg.Scene.TraceDepth {
			if material.ReflectionCol > 0.0 { // -------- Reflexion
				T := originBackV.Dot(vNormal)
				if T > 0.0 {
					vDirRef := (vNormal.Mul(2).Mul(T)).Sub(originBackV)
					vOffsetInter := interPoint.Add(vDirRef.Mul(SMALL))
					rayoRef := NewRay(vOffsetInter, vDirRef)
					c = c.Add(rg.trace(rayoRef, depth+1.0).Mul(material.ReflectionCol))
				}
			}
			if material.TransmitCol > 0.0 { // ---- Refraccion
				RN := vNormal.Dot(incidentV.Mul(-1.0))
				incidentV = incidentV.Normalize()
				var n1, n2 float64
				if vNormal.Dot(incidentV) > 0.0 {
					vNormal = vNormal.Mul(-1.0)
					RN = -RN
					n1 = material.IOR
					n2 = 1.0
				} else {
					n2 = material.IOR
					n1 = 1.0
				}
				if n1 != 0.0 && n2 != 0.0 {
					par_sqrt := math.Sqrt(1 - (n1*n1/n2*n2)*(1-RN*RN))
					refactDirV := incidentV.Add(vNormal.Mul(RN).Mul(n1 / n2)).Sub(vNormal.Mul(par_sqrt))
					vOffsetInter := interPoint.Add(refactDirV.Mul(SMALL))
					refractRay := NewRay(vOffsetInter, refactDirV)
					c = c.Add(rg.trace(refractRay, depth+1.0).Mul(material.TransmitCol))
				}
			}
		}
	}
	return c
}

func (rg *RayGun) renderPixel(line chan int, done chan bool) {
	for y := range line { // 1: 1, 5: 2, 8: 3,
		for x := 0; x < rg.Scene.ImgWidth; x++ {
			var c Color
			yo := y * rg.Scene.OverSampling
			xo := x * rg.Scene.OverSampling
			for i := 0; i < rg.Scene.OverSampling; i++ {
				for j := 0; j < rg.Scene.OverSampling; j++ {
					dir := &Vector{}
					dir.X = float64(xo)*rg.Scene.Vhor.X + float64(yo)*rg.Scene.Vver.X + rg.Scene.Vp.X
					dir.Y = float64(xo)*rg.Scene.Vhor.Y + float64(yo)*rg.Scene.Vver.Y + rg.Scene.Vp.Y
					dir.Z = float64(xo)*rg.Scene.Vhor.Z + float64(yo)*rg.Scene.Vver.Z + rg.Scene.Vp.Z
					dir = dir.Normalize()
					r := NewRay(rg.Scene.CameraPos, dir)
					c = c.Add(rg.trace(r, 1))
					yo += 1
				}
				xo += 1
			}
			srq_oversampling := float64(rg.Scene.OverSampling * rg.Scene.OverSampling)
			c.R /= srq_oversampling
			c.G /= srq_oversampling
			c.B /= srq_oversampling
			rg.Scene.Image.SetRGBA(x, y, c.ToPixel())
		}
		if y%100 == 0 {
			fmt.Printf("%d ", y)
		}
	}
	done <- true
}

func (rg *RayGun) Write() {
	reset := func(buffer *bytes.Buffer) {
		buffer.Reset()
		buffer.WriteString("package components\n\n")
		buffer.WriteString("import (\n\t\"github.com/IMQS/raygun\"\n)\n\n")
	}
	path := "C:/Projects/siteview/src/github.com/IMQS/siteview/components/"
	buffer := bytes.NewBufferString("")
	for _, group := range rg.Scene.GroupList {
		reset(buffer)
		buffer.WriteString(fmt.Sprintf("var %s = ", group.Name))
		fmt.Fprintf(buffer, "%#v\n", group.Name, group)
		ioutil.WriteFile(path+group.Name+".go", []byte(buffer.String()), os.ModePerm)
	}

	rg.Scene.GroupList = nil
	rg.Scene.Image = nil
	reset(buffer)
	buffer.WriteString("var scene = ")
	fmt.Fprintf(buffer, "%#v", rg.Scene)
	ioutil.WriteFile(path+"scene.go", []byte(buffer.String()), os.ModePerm)
	// for _, group := range rg.Scene.GroupList {
	// 	group.Write(buffer)
	// }
}
