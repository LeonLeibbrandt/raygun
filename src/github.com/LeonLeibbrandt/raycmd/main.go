package main

import (
	// "bufio"
	"flag"
	"fmt"
	"github.com/LeonLeibbrandt/raygun"
	// "os"
	"runtime"
	"time"
)

func main() {
	start := time.Now()
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
	// rg.Write()
	taken := time.Since(start)
	fmt.Printf("Time taken : %s for %v objects\n", taken, rg.Scene.ObjectCount())
}
