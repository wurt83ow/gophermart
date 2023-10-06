// workerpoo/pool.go

package workerpool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wurt83ow/gophermart/internal/models"
)

type External interface {
	GetExtOrderAccruel(string) (models.ExtRespOrder, error)
}
type Storage interface {
	GetOpenOrders() ([]string, error)
	UpdateOrderStatus([]models.ExtRespOrder) error
	InsertAccruel(map[string]models.ExtRespOrder) error
}

// Pool воркера
type Pool struct {
	Tasks   []*Task
	Workers []*Worker

	concurrency   int
	collector     chan *Task
	runBackground chan bool
	results       chan interface{}
	wg            sync.WaitGroup
	cancelFunc    context.CancelFunc
	external      External
	storage       Storage
}

// NewPool инициализирует новый пул с заданными задачами и

func NewPool(tasks []*Task, concurrency int, external External, storage Storage) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		collector:   make(chan *Task, 1000),
		results:     make(chan interface{}, 1000),
		external:    external,
		storage:     storage,
	}
}

// Run запускает всю работу в Pool и блокирует ее до тех пор,
// пока она не будет закончена.
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

// AddTask добавляет таски в pool
func (p *Pool) AddTask(task *Task) {
	p.collector <- task
}

// AddResults добавляет result в pool
func (p *Pool) AddResults(result interface{}) {
	p.results <- result
}

func (p *Pool) GetResults() <-chan interface{} {
	// close(p.results)
	return p.results
}

// RunBackground запускает pool в фоне
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

	ctx := context.Background()
	ctx, canselFunc := context.WithCancel(ctx)
	p.cancelFunc = canselFunc
	p.wg.Add(1)
	go p.UpdateOrders(ctx)

	p.runBackground = make(chan bool)
	<-p.runBackground
}

// Stop останавливает запущенных в фоне worker-ов
func (p *Pool) Stop() {
	for i := range p.Workers {
		p.Workers[i].Stop()
	}

	p.cancelFunc()
	p.wg.Wait()

	p.runBackground <- true
}

func (p *Pool) UpdateOrders(ctx context.Context) {

	t := time.NewTicker(3 * time.Second)

	result := make([]models.ExtRespOrder, 0)

	var dmx sync.RWMutex
	dmx.RLock()
	defer dmx.RUnlock()

	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.results:

			result = append(result, job.(models.ExtRespOrder))

		case <-t.C:
			orders, err := p.storage.GetOpenOrders()
			if err != nil {
				return
			}
			p.CreateOrdersTask(orders)

			if len(result) != 0 {
				p.doWork(result)
				result = nil
			}
		}

	}
}

func (p *Pool) CreateOrdersTask(orders []string) {
	var task *Task

	for _, o := range orders {
		taskID := o
		task = NewTask(func(data interface{}) error {
			order := data.(string)
			orderdata, err := p.external.GetExtOrderAccruel(order)

			if err != nil {
				return err
			}

			fmt.Printf("Task %s processed\n", order)
			p.AddResults(orderdata)
			return nil
		}, taskID)
		p.AddTask(task)
	}
}

func (p *Pool) doWork(result []models.ExtRespOrder) {

	// 1.Вызвать методы storage UpdateOrderStatus и метод кипера
	// для группового обновления таблицы orders (поле статус)
	err := p.storage.UpdateOrderStatus(result)
	if err != nil {
		//!!! перенести лог и сообщить что-то
		fmt.Println("7777777777777777777777777777777777", err)
	}

	// 2. Отобрать в массиве только структуры с accruel и вызвать
	// метод storage InsertAccruel и метод кипера для добавления записей
	// в savings_account

	var dmx sync.RWMutex

	//!!! Здесь оставить только записи с полем accruel
	orders := make(map[string]models.ExtRespOrder, 0)
	for _, o := range result {
		if o.Accrual != 0 {
			dmx.RLock()
			orders[o.Order] = o
			dmx.RUnlock()
		}
	}

	err = p.storage.InsertAccruel(orders)
	if err != nil {
		//!!! перенести лог и сообщить что-то
		fmt.Println("7777777777777777777777777777777777", err)
	}

}
