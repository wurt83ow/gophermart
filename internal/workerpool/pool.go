package workerpool

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/wurt83ow/gophermart/internal/models"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type External interface {
	GetExtOrderAccruel(string) (models.ExtRespOrder, error)
}

type Log interface {
	Info(string, ...zapcore.Field)
}

// Pool
type Pool struct {
	Tasks   []*Task
	Workers []*Worker

	concurrency   int
	collector     chan *Task
	runBackground chan bool
	wg            sync.WaitGroup
	log           Log
}

// NewPool initializes a new pool with the given tasks
func NewPool(tasks []*Task, concurrency func() string, log Log) *Pool {

	conc, err := strconv.Atoi(concurrency())
	if err != nil {
		log.Info("cannot convert concurrency option: ", zap.Error(err))
		conc = 5
	}

	return &Pool{
		Tasks:       tasks,
		concurrency: conc,
		collector:   make(chan *Task, 1000),
		log:         log,
	}
}

// Starts all the work in the Pool and blocks until it is finished.
func (p *Pool) Run() {
	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		worker.Start(&p.wg)
	}

	for i := range p.Tasks {
		p.collector <- p.Tasks[i]
	}
	close(p.collector)

	p.wg.Wait()
}

// AddTask adds tasks to the pool
func (p *Pool) AddTask(task *Task) {
	p.collector <- task
}

// RunBackground runs the pool in the background
func (p *Pool) RunBackground() {
	go func() {
		for {
			fmt.Print("⌛ Waiting for tasks to come in ...\n")
			time.Sleep(3 * time.Second)
		}
	}()

	for i := 1; i <= p.concurrency; i++ {
		worker := NewWorker(p.collector, i)
		p.Workers = append(p.Workers, worker)
		go worker.StartBackground()
	}

	for i := range p.Tasks {
		p.collector <- p.Tasks[i]
	}

	p.runBackground = make(chan bool)
	<-p.runBackground
}

// Stop stops workers running in the background
func (p *Pool) Stop() {
	for i := range p.Workers {
		p.Workers[i].Stop()
	}

	// p.cancelFunc()
	// p.wg.Wait()

	p.runBackground <- true
}

// func (p *Pool) UpdateOrders(ctx context.Context) {

// 	t := time.NewTicker(3 * time.Second)

// 	result := make([]models.ExtRespOrder, 0)

// 	var dmx sync.RWMutex
// 	dmx.RLock()
// 	defer dmx.RUnlock()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case job := <-p.results:
// 			result = append(result, job.(models.ExtRespOrder))
// 		case <-t.C:
// 			orders, err := p.storage.GetOpenOrders()
// 			if err != nil {
// 				return
// 			}
// 			p.CreateOrdersTask(orders)

// 			if len(result) != 0 {
// 				p.doWork(result)
// 				result = nil
// 			}
// 		}
// 	}
// }

// func (p *Pool) CreateOrdersTask(orders []string) {
// 	var task *Task

// 	for _, o := range orders {
// 		taskID := o
// 		task = NewTask(func(data interface{}) error {
// 			order := data.(string)
// 			orderdata, err := p.external.GetExtOrderAccruel(order)

// 			if err != nil {
// 				return err
// 			}

// 			fmt.Printf("Task %s processed\n", order)
// 			p.AddResults(orderdata)
// 			return nil
// 		}, taskID)
// 		p.AddTask(task)
// 	}
// }

// func (p *Pool) doWork(result []models.ExtRespOrder) {

// 	// perform a group update of the orders table (status field)
// 	err := p.storage.UpdateOrderStatus(result)
// 	if err != nil {
// 		p.log.Info("errors when updating order status: ", zap.Error(err))
// 	}

// 	// add records with accruel to savings_account
// 	var dmx sync.RWMutex

// 	orders := make(map[string]models.ExtRespOrder, 0)
// 	for _, o := range result {
// 		if o.Accrual != 0 {
// 			dmx.RLock()
// 			orders[o.Order] = o
// 			dmx.RUnlock()
// 		}
// 	}

// 	err = p.storage.InsertAccruel(orders)
// 	if err != nil {
// 		p.log.Info("errors when accruel inserting: ", zap.Error(err))
// 	}

// }
