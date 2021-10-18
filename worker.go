package xiaolongbaopool

import (
	"log"
	"time"
)

type Worker struct {
	pool *Pool

	task chan func()

	recycleTime time.Time
}

func (w *Worker) run() {
	w.pool.incRunning()

	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
				log.Printf("worker exits from a panic: %v", p)
			}
		}()

		for f := range w.task {
			if f == nil {
				w.pool.decRunning()
				w.pool.workerCache.Put(w)
				return
			}
			f()
			if err := w.pool.putWorker(w); err != nil {
				break
			}
		}
	}()
}
