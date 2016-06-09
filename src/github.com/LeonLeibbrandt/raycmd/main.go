package main

import (
	"flag"
	"github.com/LeonLeibbrandt/raygun"
	"runtime"
	"fmt"
	"os"
)

func main() {
	scenefile := flag.String("scene", "", "Scene file to parse")
	numcpu := flag.Int("numcpu", 0, "Number of cores to use, default to available")
	flag.Parse()

	if *scenefile == "" {
		fmt.Println("Usage raygun --scene <scenefile>")
		os.Exit(-1)
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
