package main

import (
	"flag"
	"fmt"
	"github.com/LeonLeibbrandt/raygun"
	"os"
	"runtime"
)

func main() {
	scenefile := flag.String("scene", "", "Scene file")
	numcpu := flag.Int("numcpu", 0, "Number Of cores to use")
	flag.Parse()

	if *scenefile == "" {
		fmt.Println("Usage: raygun --scene path/to/scene.txt --numcpu [Number of cores; defaults to all]")
		os.Exit(0)
	}
	if *numcpu == 0 {
		*numcpu = runtime.NumCPU()
	}

	rg, err := raygun.NewRayGun(*scenefile, *numcpu)
	if err != nil {
		panic(err)
	}

	rg.Render()
}
