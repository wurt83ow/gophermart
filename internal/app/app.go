package app

import (
	"errors"
	"fmt"

	"github.com/wurt83ow/gophermart/internal/config"
	"github.com/wurt83ow/gophermart/internal/logger"
)

func Run() error {
	option := config.NewOptions()
	option.ParseFlags()

	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {
		return err
	}
	fmt.Println(nLogger)
	return errors.New("test")
}
