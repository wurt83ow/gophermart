package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

type Log interface {
	Warn(msg string, fields ...zapcore.Field)
	Info(string, ...zapcore.Field)
}

type Storage interface {
	// DeleteURLs(delUrls ...models.DeleteURL) error
}

type ExtController interface {
	// DeleteURLs(delUrls ...models.DeleteURL) error
}

type worker struct {
	wg         *sync.WaitGroup
	cancelFunc context.CancelFunc
	log        Log
	storage    Storage
	extcontr   ExtController
	// jobChan    chan models.DeleteURL
	// result     []models.DeleteURL //chan interface{}
}

// type workType string
type Worker interface {
	Start(pctx context.Context)
	Stop()
	// Add(models.DeleteURL)
}

func NewWorker(extcontr ExtController, storage Storage, log Log) Worker {
	w := worker{
		wg:       new(sync.WaitGroup),
		log:      log,
		storage:  storage,
		extcontr: extcontr,
		// jobChan: make(chan models.DeleteURL, 1024),
		// result:  make([]models.DeleteURL, 0),
	}

	return &w
}

func (w *worker) Start(pctx context.Context) {
	w.log.Warn("Start worker")
	ctx, canselFunc := context.WithCancel(pctx)
	w.cancelFunc = canselFunc
	w.wg.Add(1)
	go w.spawnWorkers(ctx)
}

func (w *worker) Stop() {
	w.log.Warn("Stop worker")
	w.cancelFunc()
	w.wg.Wait()
	w.log.Warn("All workers exited!")
}

// func (w *worker) Add(d models.DeleteURL) {
// 	w.jobChan <- d
// }

func (w *worker) spawnWorkers(ctx context.Context) {
	defer w.wg.Done()
	w.log.Warn(" start ")
	t := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		// case job := <-w.jobChan:
		// 	w.result = append(w.result, job)
		case <-t.C:
			w.doWork(ctx)
		}
	}
}

func (w *worker) doWork(ctx context.Context) {
	// if len(w.result) != 0 {
	// 	err := w.storage.DeleteURLs(w.result...)
	// 	if err != nil {
	// 		w.log.Info("cannot save delUrls", zap.Error(err))
	// 	}
	// 	w.result = nil
	// }
}
