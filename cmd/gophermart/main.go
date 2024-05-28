package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/wurt83ow/gophermart/internal/app"
)

func main() {
	// Создание файла для записи профиля CPU
	fl, err := os.Create("./cpu.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer fl.Close()

	// Запуск профилирования CPU
	if err := pprof.StartCPUProfile(fl); err != nil {
		//nolint:gocritic
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Создание корневого контекста с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())

	// Создание канала для обработки сигналов
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Запуск сервера
	server := app.NewServer(ctx)
	go func() {
		// Ожидание сигнала
		sig := <-signalCh
		log.Printf("Received signal: %+v", sig)

		// Завершение работы сервера
		server.Shutdown()

		// Отмена контекста
		cancel()

		// Закрытие файла и завершение профилирования
		fl.Close()
	}()

	// Запуск сервера
	server.Serve()
}
