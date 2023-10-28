package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/wurt83ow/gophermart/internal/app"
)

func main() {
	fl, err := os.Create("./cpu.pprof")
	if err != nil {
		log.Fatal()
	}
	defer fl.Close()

	pprof.StartCPUProfile(fl)
	defer pprof.StopCPUProfile()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
