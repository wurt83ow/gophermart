package main

import "github.com/wurt83ow/gophermart/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
