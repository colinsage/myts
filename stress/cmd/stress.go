package main

import (
	"flag"
	"os"
	"runtime/pprof"
	"fmt"
	"github.com/colinsage/myts/stress"
)

var (
	config     = flag.String("config", "", "The stress test file")
	cpuprofile = flag.String("cpuprofile", "", "Write the cpu profile to `filename`")
	db         = flag.String("db", "", "target database within test system for write and query load")
)

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *config != "" {
		stress.RunStress(*config)
	} else {
		stress.RunStress("stress/iql/file.iql")
	}

}
