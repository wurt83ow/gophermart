package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

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
	// Создание файла для записи профиля CPU
	fl, err := os.Create("./cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	// defer fl.Close()

<<<<<<< HEAD
	// Запуск профилирования CPU
	if err := pprof.StartCPUProfile(fl); err != nil {
		//nolint:gocritic
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()
=======
	pprof.StartCPUProfile(fl)
	// defer pprof.StopCPUProfile()

	n := fib(40)
>>>>>>> ef554a6343e21ec465ecff19742d925a0f910ef9

	// Создание корневого контекста с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())

	// Создание канала для обработки сигналов
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера
	server := app.NewServer(ctx)
	go func() {
<<<<<<< HEAD
		// Ожидание сигнала
		sig := <-signalCh
		log.Printf("Received signal: %+v", sig)

		// Завершение работы сервера
		server.Shutdown()

		// Отмена контекста
=======
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		fl.Close()
		pprof.StopCPUProfile()
		fmt.Println("8888888888888888888888888888888888888", n)
		server.Shutdown()

>>>>>>> ef554a6343e21ec465ecff19742d925a0f910ef9
		cancel()

		// Закрытие файла и завершение профилирования
		fl.Close()
	}()

	// Запуск сервера
	server.Serve()
}
