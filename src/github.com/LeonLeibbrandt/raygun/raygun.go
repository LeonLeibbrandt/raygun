package raygun

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"math"
	"os"
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
		Scene:      NewScene(filename),
		Done:       make(chan bool, numworkers),
		Line:       make(chan int),
	}
	return rg, nil
}

func (rg *RayGun) Render() {
	for i := 0; i < rg.NumWorkers; i++ {
		go rg.renderPixel(rg.Line, rg.Done)
	}

	fmt.Println("Rendering: ", rg.FileName)
	fmt.Printf("Line (from %d to %d): ", rg.Scene.StartLine, rg.Scene.EndLine)

	for y := rg.Scene.StartLine; y < rg.Scene.EndLine; y++ {
		rg.Line <- y
	}
	close(rg.Line)

	// wait for all workers to finish
	for i := 0; i < rg.NumWorkers; i++ {
		<-rg.Done
	}

	output, err := os.Create(rg.FileName + ".jpg")
	if err != nil {
		panic(err)
	}

	err = jpeg.Encode(output, rg.Scene.Image, nil)
	// err = png.Encode(output, rg.Scene.Image)
	if err != nil {
		panic(err)
	}
	fmt.Println("DONE!")

}

func (rg *RayGun) calcShadow(r *Ray, collisionObj, collisionGrp int) float64 {
	shadow := 1.0 //starts with no shadow
	for g, grp := range rg.Scene.GroupList {
		for i, obj := range grp.ObjectList {
			r.interObj = -1
			r.interGrp = -1
			r.interDist = MAX_DIST

			if obj.Intersect(r, g, i) && g != collisionGrp && i != collisionObj {
				shadow *= rg.Scene.MaterialList[obj.Material()].TransmitCol
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
			obj.Intersect(r, g, i)
		}
	}

	if r.interObj >= 0 {
		matIndex := rg.Scene.GroupList[r.interGrp].ObjectList[r.interObj].Material()
		interPoint := r.origin.Add(r.direction.Mul(r.interDist))
		incidentV := interPoint.Sub(r.origin)
		originBackV := r.direction.Mul(-1.0)
		originBackV = originBackV.Normalize()
		vNormal := rg.Scene.GroupList[r.interGrp].ObjectList[r.interObj].getNormal(interPoint)
		for _, light := range rg.Scene.LightList {
			switch light.Kind {
			case "ambient":
				c = c.Add(light.Color)
			case "point":
				lightDir := light.Position.Sub(interPoint)
				lightDir = lightDir.Normalize()
				lightRay := NewRay(interPoint, lightDir)
				shadow := rg.calcShadow(lightRay, r.interObj, r.interGrp)
				NL := vNormal.Dot(lightDir)

				if NL > 0.0 {
					if rg.Scene.MaterialList[matIndex].DifuseCol > 0.0 { // ------- Difuso
						difuseColor := light.Color.Mul(rg.Scene.MaterialList[matIndex].DifuseCol).Mul(NL)
						difuseColor.R *= rg.Scene.MaterialList[matIndex].Color.R  * shadow
						difuseColor.G *= rg.Scene.MaterialList[matIndex].Color.G  * shadow
						difuseColor.B *= rg.Scene.MaterialList[matIndex].Color.B  * shadow
						c = c.Add(difuseColor)
					}
					if rg.Scene.MaterialList[matIndex].SpecularCol > 0.0 { // ----- Especular
						R := (vNormal.Mul(2).Mul(NL)).Sub(lightDir)
						spec := originBackV.Dot(R)
						if spec > 0.0 {
							spec = rg.Scene.MaterialList[matIndex].SpecularCol * math.Pow(spec, rg.Scene.MaterialList[matIndex].SpecularD)
							specularColor := light.Color.Mul(spec).Mul(shadow)
							c = c.Add(specularColor)
						}
					}
				}
			}
		}
		if depth < rg.Scene.TraceDepth {
			if rg.Scene.MaterialList[matIndex].ReflectionCol > 0.0 { // -------- Reflexion
				T := originBackV.Dot(vNormal)
				if T > 0.0 {
					vDirRef := (vNormal.Mul(2).Mul(T)).Sub(originBackV)
					vOffsetInter := interPoint.Add(vDirRef.Mul(SMALL))
					rayoRef := NewRay(vOffsetInter, vDirRef)
					c = c.Add(rg.trace(rayoRef, depth+1.0).Mul(rg.Scene.MaterialList[matIndex].ReflectionCol))
				}
			}
			if rg.Scene.MaterialList[matIndex].TransmitCol > 0.0 { // ---- Refraccion
				RN := vNormal.Dot(incidentV.Mul(-1.0))
				incidentV = incidentV.Normalize()
				var n1, n2 float64
				if vNormal.Dot(incidentV) > 0.0 {
					vNormal = vNormal.Mul(-1.0)
					RN = -RN
					n1 = rg.Scene.MaterialList[matIndex].IOR
					n2 = 1.0
				} else {
					n2 = rg.Scene.MaterialList[matIndex].IOR
					n1 = 1.0
				}
				if n1 != 0.0 && n2 != 0.0 {
					par_sqrt := math.Sqrt(1 - (n1*n1/n2*n2)*(1-RN*RN))
					refactDirV := incidentV.Add(vNormal.Mul(RN).Mul(n1 / n2)).Sub(vNormal.Mul(par_sqrt))
					vOffsetInter := interPoint.Add(refactDirV.Mul(SMALL))
					refractRay := NewRay(vOffsetInter, refactDirV)
					c = c.Add(rg.trace(refractRay, depth+1.0).Mul(rg.Scene.MaterialList[matIndex].TransmitCol))
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
					c = c.Add(rg.trace(r, 1.0))
					yo += 1
				}
				xo += 1
			}
			srq_oversampling := float64(rg.Scene.OverSampling * rg.Scene.OverSampling)
			c.R /= srq_oversampling
			c.G /= srq_oversampling
			c.B /= srq_oversampling
			rg.Scene.Image.SetRGBA(x, y, c.ToPixel())
			//fmt.Println("check")
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
