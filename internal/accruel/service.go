package accruel

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/wurt83ow/gophermart/internal/models"
	"github.com/wurt83ow/gophermart/internal/workerpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type External interface {
	GetExtOrderAccruel(string) (models.ExtRespOrder, error)
}

type Log interface {
	Info(string, ...zapcore.Field)
}

type Storage interface {
	GetOpenOrders() ([]string, error)
	UpdateOrderStatus([]models.ExtRespOrder) error
	InsertAccruel(map[string]models.ExtRespOrder) error
}

type Pool interface {
	// NewTask(f func(interface{}) error, data interface{}) *workerpool.Task
	AddTask(task *workerpool.Task)
}

type AccrualService struct {
	results      chan interface{}
	wg           sync.WaitGroup
	cancelFunc   context.CancelFunc
	external     External
	pool         Pool
	storage      Storage
	log          Log
	taskInterval int
}

func NewAccrualService(external External, pool Pool, storage Storage, log Log, TaskExecutionInterval func() string) *AccrualService {
	taskInterval, err := strconv.Atoi(TaskExecutionInterval())
	if err != nil {
		log.Info("cannot convert concurrency option: ", zap.Error(err))
		taskInterval = 3000
	}

	return &AccrualService{
		results:      make(chan interface{}, 1000),
		external:     external,
		pool:         pool,
		storage:      storage,
		log:          log,
		taskInterval: taskInterval,
	}
}

// starts a worker.
func (a *AccrualService) Start() {
	ctx := context.Background()
	ctx, canselFunc := context.WithCancel(ctx)
	a.cancelFunc = canselFunc
	a.wg.Add(1)
	go a.UpdateOrders(ctx)
}

func (a *AccrualService) Stop() {
	a.cancelFunc()
	a.wg.Wait()
}

func (a *AccrualService) UpdateOrders(ctx context.Context) {
	t := time.NewTicker(time.Duration(a.taskInterval) * time.Millisecond)

	result := make([]models.ExtRespOrder, 0)

	var dmx sync.RWMutex
	dmx.RLock()
	defer dmx.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-a.results:
			result = append(result, job.(models.ExtRespOrder))
		case <-t.C:
			orders, err := a.storage.GetOpenOrders()
			if err != nil {
				return
			}
			a.CreateOrdersTask(orders)

			if len(result) != 0 {
				a.doWork(result)
				result = nil
			}
		}
	}
}

// AddResults adds result to pool.
func (a *AccrualService) AddResults(result interface{}) {
	a.results <- result
}

func (a *AccrualService) GetResults() <-chan interface{} {
	// close(p.results)
	return a.results
}

func (a *AccrualService) CreateOrdersTask(orders []string) {
	var task *workerpool.Task

	for _, o := range orders {
		taskID := o
		task = workerpool.NewTask(func(data interface{}) error {
			order := data.(string)
			orderdata, err := a.external.GetExtOrderAccruel(order)
			if err != nil {
				return err
			}

			fmt.Printf("Task %s processed\n", order)
			a.AddResults(orderdata)
			return nil
		}, taskID)
		a.pool.AddTask(task)
	}
}

func (a *AccrualService) doWork(result []models.ExtRespOrder) {
	// perform a group update of the orders table (status field)
	err := a.storage.UpdateOrderStatus(result)
	if err != nil {
		a.log.Info("errors when updating order status: ", zap.Error(err))
	}

	// add records with accruel to savings_account
	var dmx sync.RWMutex

	orders := make(map[string]models.ExtRespOrder, 0)
	for _, o := range result {
		if o.Accrual != 0 {
			dmx.RLock()
			orders[o.Order] = o
			dmx.RUnlock()
		}
	}

	err = a.storage.InsertAccruel(orders)
	if err != nil {
		a.log.Info("errors when accruel inserting: ", zap.Error(err))
	}
}
