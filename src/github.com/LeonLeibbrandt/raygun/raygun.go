package raygun

import (
	"bufio"
	"fmt"
	"image/png"
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

	output, err := os.Create(rg.FileName + ".png")
	if err != nil {
		panic(err)
	}

	err = png.Encode(output, rg.Scene.image)
	if err != nil {
		panic(err)
	}
	fmt.Println(" DONE!")

}

func (rg *RayGun) calcShadow(r *Ray, collisionObj int) float64 {
	shadow := 1.0 //starts with no shadow
	for i, obj := range rg.Scene.objectList {
		r.interObj = -1
		r.interDist = MAX_DIST

		if obj.Intersect(r, i) && i != collisionObj {
			shadow *= rg.Scene.materialList[obj.MaterialIndex()].transmitCol
		}
	}
	return shadow
}

func (rg *RayGun) trace(r *Ray, depth int) (c Color) {

	for i, obj := range rg.Scene.objectList {
		obj.Intersect(r, i)
	}

	if r.interObj >= 0 {
		matIndex := rg.Scene.objectList[r.interObj].MaterialIndex()
		interPoint := r.origin.Add(r.direction.Mul(r.interDist))
		incidentV := interPoint.Sub(r.origin)
		originBackV := r.direction.Mul(-1.0)
		originBackV = originBackV.Normalize()
		vNormal := rg.Scene.objectList[r.interObj].getNormal(interPoint)
		for _, light := range rg.Scene.lightList {
			switch light.kind {
			case "ambient":
				c = c.Add(light.color)
			case "point":
				lightDir := light.position.Sub(interPoint)
				lightDir = lightDir.Normalize()
				lightRay := Ray{interPoint, lightDir, MAX_DIST, -1}
				shadow := rg.calcShadow(&lightRay, r.interObj)
				NL := vNormal.Dot(lightDir)

				if NL > 0.0 {
					if rg.Scene.materialList[matIndex].difuseCol > 0.0 { // ------- Difuso
						difuseColor := light.color.Mul(rg.Scene.materialList[matIndex].difuseCol).Mul(NL)
						difuseColor.r *= rg.Scene.materialList[matIndex].color.r * shadow
						difuseColor.g *= rg.Scene.materialList[matIndex].color.g * shadow
						difuseColor.b *= rg.Scene.materialList[matIndex].color.b * shadow
						c = c.Add(difuseColor)
					}
					if rg.Scene.materialList[matIndex].specularCol > 0.0 { // ----- Especular
						R := (vNormal.Mul(2).Mul(NL)).Sub(lightDir)
						spec := originBackV.Dot(R)
						if spec > 0.0 {
							spec = rg.Scene.materialList[matIndex].specularCol * math.Pow(spec, rg.Scene.materialList[matIndex].specularD)
							specularColor := light.color.Mul(spec).Mul(shadow)
							c = c.Add(specularColor)
						}
					}
				}
			}
		}
		if depth < rg.Scene.traceDepth {
			if rg.Scene.materialList[matIndex].reflectionCol > 0.0 { // -------- Reflexion
				T := originBackV.Dot(vNormal)
				if T > 0.0 {
					vDirRef := (vNormal.Mul(2).Mul(T)).Sub(originBackV)
					vOffsetInter := interPoint.Add(vDirRef.Mul(SMALL))
					rayoRef := Ray{vOffsetInter, vDirRef, MAX_DIST, -1}
					c = c.Add(rg.trace(&rayoRef, depth+1.0).Mul(rg.Scene.materialList[matIndex].reflectionCol))
				}
			}
			if rg.Scene.materialList[matIndex].transmitCol > 0.0 { // ---- Refraccion
				RN := vNormal.Dot(incidentV.Mul(-1.0))
				incidentV = incidentV.Normalize()
				var n1, n2 float64
				if vNormal.Dot(incidentV) > 0.0 {
					vNormal = vNormal.Mul(-1.0)
					RN = -RN
					n1 = rg.Scene.materialList[matIndex].IOR
					n2 = 1.0
				} else {
					n2 = rg.Scene.materialList[matIndex].IOR
					n1 = 1.0
				}
				if n1 != 0.0 && n2 != 0.0 {
					par_sqrt := math.Sqrt(1 - (n1*n1/n2*n2)*(1-RN*RN))
					refactDirV := incidentV.Add(vNormal.Mul(RN).Mul(n1 / n2)).Sub(vNormal.Mul(par_sqrt))
					vOffsetInter := interPoint.Add(refactDirV.Mul(SMALL))
					refractRay := Ray{vOffsetInter, refactDirV, MAX_DIST, -1}
					c = c.Add(rg.trace(&refractRay, depth+1.0).Mul(rg.Scene.materialList[matIndex].transmitCol))
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
					r := Ray{rg.Scene.cameraPos, dir, MAX_DIST, -1}
					c = c.Add(rg.trace(&r, 1.0))
					yo += 1
				}
				xo += 1
			}
			srq_oversampling := float64(rg.Scene.oversampling * rg.Scene.oversampling)
			c.r /= srq_oversampling
			c.g /= srq_oversampling
			c.b /= srq_oversampling
			rg.Scene.image.SetRGBA(x, y, c.ToPixel())
			//fmt.Println("check")
		}
		if y%100 == 0 {
			fmt.Printf("%d ", y)
		}
	}
	done <- true
}

func (rg *RayGun) Write(buffer *bufio.Writer) {
	for _, object := range rg.Scene.objectList {
		object.Write(buffer)
	}
}
