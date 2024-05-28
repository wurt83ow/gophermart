package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wurt83ow/gophermart/internal/models"
	"go.uber.org/zap"
)

type ExtController struct {
	storage Storage
	log     Log
	extAddr func() string
}

type Pool interface {
	AddResults(interface{})
	GetResults() <-chan interface{}
}

func NewExtController(storage Storage, extAddr func() string, log Log) *ExtController {
	return &ExtController{
		storage: storage,
		log:     log,
		extAddr: extAddr,
	}
}

func (c *ExtController) GetExtOrderAccruel(order string) (models.ExtRespOrder, error) {
	addr := c.extAddr()
	if string(addr[len(addr)-1]) != "/" {
		addr = addr + "/"
	}

	url := addr + "api/orders/" + order

	resp, err := http.Get(url)
	if err != nil {
		c.log.Info("unable to access accruel service, check that it is running: ", zap.Error(err))
		return models.ExtRespOrder{}, err // Добавлено возвращение ошибки
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Info("status code error: ", zap.String("method", resp.Status))
		return models.ExtRespOrder{}, fmt.Errorf("status code error: %s", resp.Status) // Добавлено возвращение ошибки
	}

	respOrd := models.ExtRespOrder{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respOrd); err != nil {
		return models.ExtRespOrder{}, err
	}

	return respOrd, nil
}
