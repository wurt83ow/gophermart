package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/workerpool"
)

type ExtController struct {
	storage Storage
	log     Log
}

type Pool interface {
	AddTask(*workerpool.Task)
	AddResults(interface{})
	GetResults() <-chan interface{}
}

func NewExtController(storage Storage, log Log) *ExtController {
	return &ExtController{
		storage: storage,
		log:     log,
	}
}

func (c *ExtController) GetOrder(order string) (models.ExtRespOrder, error) {

	url := "http://localhost:8082/api/orders/" + order // !!! прокинуть config

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err) //!!! log
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("status code error: ", resp.StatusCode, resp.Status)
	}

	respOrd := models.ExtRespOrder{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&respOrd); err != nil {
		return models.ExtRespOrder{}, err
	}

	return respOrd, nil
}

// func (c *ExtController) ResultProcessing() {

// 	t := time.NewTicker(10 * time.Second)
// 	result := make([]interface{}, 0)
// 	for {
// 		select {
// 		case job := <-c.pool.GetResults():
// 			result = append(result, job)
// 		case <-t.C:
// 			if len(result) != 0 {
// 				//!!! Создать и загрузить результат в массив структур
// 				// 1.Вызвать методы  storage UpdateOrderStatus и метод кипера
// 				// для группового обновления таблицы orders (поле статус)
// 				// 2. Отобрать в массиве только структуры с accruel и вызвать
// 				// метод storage InsertAccruel и метод кипера для добавления записей
// 				// в savings_account

// 				result = nil
// 			}
// 		}
// 	}
// }
