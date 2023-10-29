package main

import (
	"context"
	"log"
	"os"
	"os/signal"
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

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	server := app.NewServer(ctx)

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		server.Shutdown()
		cancel()
	}()

	server.Serve()
}
