package main

import (
	"image/png"
	//	"encoding/json"
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

	/*
		jsonfile := *scenefile + ".json"
		buf, _ := json.MarshalIndent(rg.Scene, "", "\t")
		file, _ := os.Create(jsonfile)
		file.Write(buf)
		file.Close()
	*/
	rg.Render()

	output, err := os.Create(scenefile + ".png")
	if err != nil {
		panic(err)
	}
	defer output.Close()
	
	err = png.Encode(output, rg.Scene.Image)
	if err != nil {
		panic(err)
	}

}
