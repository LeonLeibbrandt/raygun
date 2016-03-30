package main

import (
	"flag"
	"github.com/LeonLeibbrandt/raygun"
	"runtime"
)

func main() {
	var sceneFilename string
	var numWorkers int
	flag.StringVar(&sceneFilename, "file", "samples/scene.txt", "Scene file to render.")
	flag.IntVar(&numWorkers, "workers", runtime.NumCPU(), "Number of worker threads.")
	flag.Parse()

	rg, err := raygun.NewRayGun(sceneFilename, numWorkers)
	if err != nil {
		panic(err)
	}

	rg.Render()
}
