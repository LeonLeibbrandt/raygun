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
	fmt.Printf("Line (from %d to %d): ", rg.Scene.startline, rg.Scene.endline)

	for y := rg.Scene.startline; y < rg.Scene.endline; y++ {
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
	// for g, grp := range rg.Scene.GroupList {
	//	for i, obj := range grp.ObjectList {
	//		r.interObj = -1
	//		r.interGrp = -1
	//		r.interDist = MAX_DIST
	//
	//			if obj.Intersect(r, g, i) && g != collisionGrp && i != collisionObj {
	//				shadow *= rg.Scene.MaterialList[obj.Material()].transmitCol
	//			}
	//		}
	//	}
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
			switch light.kind {
			case "ambient":
				c = c.Add(light.color)
			case "point":
				lightDir := light.position.Sub(interPoint)
				lightDir = lightDir.Normalize()
				// lightRay := Ray{interPoint, lightDir, MAX_DIST, -1, -1}
				// shadow := rg.calcShadow(&lightRay, r.interObj, r.interGrp)
				NL := vNormal.Dot(lightDir)

				if NL > 0.0 {
					if rg.Scene.MaterialList[matIndex].difuseCol > 0.0 { // ------- Difuso
						difuseColor := light.color.Mul(rg.Scene.MaterialList[matIndex].difuseCol).Mul(NL)
						difuseColor.r *= rg.Scene.MaterialList[matIndex].color.r // * shadow
						difuseColor.g *= rg.Scene.MaterialList[matIndex].color.g // * shadow
						difuseColor.b *= rg.Scene.MaterialList[matIndex].color.b // * shadow
						c = c.Add(difuseColor)
					}
					if rg.Scene.MaterialList[matIndex].specularCol > 0.0 { // ----- Especular
						R := (vNormal.Mul(2).Mul(NL)).Sub(lightDir)
						spec := originBackV.Dot(R)
						if spec > 0.0 {
							spec = rg.Scene.MaterialList[matIndex].specularCol * math.Pow(spec, rg.Scene.MaterialList[matIndex].specularD)
							specularColor := light.color.Mul(spec) // .Mul(shadow)
							c = c.Add(specularColor)
						}
					}
				}
			}
		}
		if depth < rg.Scene.traceDepth {
			if rg.Scene.MaterialList[matIndex].reflectionCol > 0.0 { // -------- Reflexion
				T := originBackV.Dot(vNormal)
				if T > 0.0 {
					vDirRef := (vNormal.Mul(2).Mul(T)).Sub(originBackV)
					vOffsetInter := interPoint.Add(vDirRef.Mul(SMALL))
					rayoRef := NewRay(vOffsetInter, vDirRef)
					c = c.Add(rg.trace(rayoRef, depth+1.0).Mul(rg.Scene.MaterialList[matIndex].reflectionCol))
				}
			}
			if rg.Scene.MaterialList[matIndex].transmitCol > 0.0 { // ---- Refraccion
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
					c = c.Add(rg.trace(refractRay, depth+1.0).Mul(rg.Scene.MaterialList[matIndex].transmitCol))
				}
			}
		}
	}
	return c
}

func (rg *RayGun) renderPixel(line chan int, done chan bool) {
	for y := range line { // 1: 1, 5: 2, 8: 3,
		for x := 0; x < rg.Scene.imgWidth; x++ {
			var c Color
			yo := y * rg.Scene.oversampling
			xo := x * rg.Scene.oversampling
			for i := 0; i < rg.Scene.oversampling; i++ {
				for j := 0; j < rg.Scene.oversampling; j++ {
					dir := &Vector{}
					dir.x = float64(xo)*rg.Scene.Vhor.x + float64(yo)*rg.Scene.Vver.x + rg.Scene.Vp.x
					dir.y = float64(xo)*rg.Scene.Vhor.y + float64(yo)*rg.Scene.Vver.y + rg.Scene.Vp.y
					dir.z = float64(xo)*rg.Scene.Vhor.z + float64(yo)*rg.Scene.Vver.z + rg.Scene.Vp.z
					dir = dir.Normalize()
					r := NewRay(rg.Scene.cameraPos, dir)
					c = c.Add(rg.trace(r, 1.0))
					yo += 1
				}
				xo += 1
			}
			srq_oversampling := float64(rg.Scene.oversampling * rg.Scene.oversampling)
			c.r /= srq_oversampling
			c.g /= srq_oversampling
			c.b /= srq_oversampling
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
	path := "/home/leon/Projects/siteview/src/github.com/IMQS/siteview/components/"
	buffer := bytes.NewBufferString("")
	buffer.WriteString("package components\n\n")
	buffer.WriteString("import (\n\t\"github.com/IMQS/raygun\"\n)\n\n")
	fmt.Fprintf(buffer, "%#v", rg.Scene)
	ioutil.WriteFile(path+"scene.go", []byte(buffer.String()), os.ModePerm)
	// for _, group := range rg.Scene.GroupList {
	// 	group.Write(buffer)
	// }
}
