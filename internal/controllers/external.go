package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/workerpool"
)

type ExtController struct {
	storage    Storage
	log        Log
	pool       Pool
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc
}

type Pool interface {
	AddTask(*workerpool.Task)
	AddResults(interface{})
	GetResults() <-chan interface{}
}

func NewExtController(storage Storage, pool Pool, log Log) *ExtController {
	return &ExtController{
		storage: storage,
		pool:    pool,
		log:     log,
		wg:      new(sync.WaitGroup),
	}
}

func (c *ExtController) Start(pctx context.Context) {
	c.log.Info("Start worker")
	ctx, canselFunc := context.WithCancel(pctx)
	c.cancelFunc = canselFunc
	c.wg.Add(1)
	go c.SubmitOrders(ctx)
}

func (c *ExtController) Stop() {
	c.log.Info("Stop worker")
	c.cancelFunc()
	c.wg.Wait()
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

func (c *ExtController) SubmitOrders(ctx context.Context) {

	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.GetOrders(ctx)
			c.ResultProcessing()
		}
	}
}

func (c *ExtController) GetOrders(ctx context.Context) {
	var task *workerpool.Task

	orders, err := c.storage.GetOpenOrders()
	if err != nil {
		return //!!! Что здесь?
	}

	for _, o := range orders {
		taskID := o
		task = workerpool.NewTask(func(data interface{}) error {
			order := data.(string)
			orderdata, err := c.GetOrder(order)

			if err != nil {
				return err
			}
			fmt.Printf("Task %s processed\n", order)
			c.pool.AddResults(orderdata)
			// time.Sleep(100 * time.Millisecond)

			return nil
		}, taskID)
		c.pool.AddTask(task)
	}
	orders = nil
}

func (c *ExtController) ResultProcessing() {

	t := time.NewTicker(10 * time.Second)
	result := make([]interface{}, 0)
	for {
		select {
		case job := <-c.pool.GetResults():
			result = append(result, job)
		case <-t.C:
			if len(result) != 0 {
				//!!! Создать и загрузить результат в массив структур
				// 1.Вызвать методы storage UpdateOrderStatus и метод кипера
				// для группового обновления таблицы orders (поле статус)
				// 2. Отобрать в массиве только структуры с accruel и вызвать
				// метод storage InsertAccruel и метод кипера для добавления записей
				// в savings_account

				result = nil
			}
		}
	}
}
