package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/wurt83ow/gophermart/internal/app"
)

func fib(n uint) uint {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return fib(n-1) + fib(n-2)
	}
}
func main() {
	fl, err := os.Create("./cpu.pprof")
	if err != nil {
		log.Fatal()
	}
	// defer fl.Close()

	pprof.StartCPUProfile(fl)
	// defer pprof.StopCPUProfile()

	n := fib(40)

	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	server := app.NewServer(ctx)

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		fl.Close()
		pprof.StopCPUProfile()
		fmt.Println("8888888888888888888888888888888888888", n)
		server.Shutdown()

		cancel()
	}()

	server.Serve()
}
