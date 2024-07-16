package controllers

import (
	"encoding/json"
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
		c.log.Info("cannot convert concurrency option: ", zap.Error(err))
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Info("status code error: ", zap.String("method", resp.Status))
	}

	respOrd := models.ExtRespOrder{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respOrd); err != nil {
		return models.ExtRespOrder{}, err
	}

	return respOrd, nil
}
