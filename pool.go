package xiaolongbaopool

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	CLOSED = 1
)

// Pool
type Pool struct {
	// capacity of the pool size
	capacity int32

	// running is the number of running worker
	running int32

	// closed specify whether the pool is closed
	closed int32

	// workers stores worker stopped
	workers []*Worker

	// workerCache is sync pool to accelerate worker creation
	workerCache sync.Pool

	// expiryDuration specify the recycle duration of the gc loop
	expiryDuration time.Duration

	lock sync.Mutex

	cond sync.Cond

	once sync.Once
}

func NewPool(size int) (*Pool, error) {
	return NewTimingPool(size, DEFAULT_CLEAN_DURATION)
}

func NewTimingPool(size, expiry int) (*Pool, error) {
	if size <= 0 {
		return nil, ErrInvalidPoolSize
	}
	if expiry <= 0 {
		return nil, ErrInvalidPoolExpiry
	}

	p := &Pool{
		capacity:       int32(size),
		expiryDuration: time.Duration(expiry) * time.Second,
	}
	p.cond = *sync.NewCond(&p.lock)

	go p.gc()

	return p, nil
}

func (p *Pool) Submit(task func()) error {
	if CLOSED == atomic.LoadInt32(&p.closed) {
		return ErrPoolClosed
	}

	p.getWorker().task <- task

	return nil
}

func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *Pool) Close() {
	p.once.Do(func() {
		atomic.StoreInt32(&p.closed, 1)
		p.lock.Lock()

		idleWorkers := p.workers
		for i, w := range idleWorkers {
			w.task <- nil
			idleWorkers[i] = nil
		}
		p.workers = nil

		p.lock.Unlock()
	})
}

func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

func (p *Pool) getWorker() *Worker {
	var w *Worker

	p.lock.Lock()
	idleWorkers := p.workers
	n := len(idleWorkers)

	if n > 0 {
		w = idleWorkers[n-1]
		idleWorkers[n-1] = nil
		p.workers = idleWorkers[:n-1]
		p.lock.Unlock()
	} else if p.Running()-p.Cap() < 0 {
		p.lock.Unlock()
		cacheWorker := p.workerCache.Get()
		if cacheWorker != nil {
			w = cacheWorker.(*Worker)
		} else {
			w = &Worker{
				pool: p,
				task: make(chan func(), 1),
			}
		}
		w.run()
	} else {
		for {
			p.cond.Wait()
			l := len(p.workers)
			if l == 0 {
				continue
			}
			w = p.workers[l-1]
			p.workers[l-1] = nil
			p.workers = p.workers[:l-1]
			break
		}
		p.lock.Unlock()
	}

	return w
}

// put back to free pool
func (p *Pool) putWorker(w *Worker) error {
	if CLOSED == atomic.LoadInt32(&p.closed) {
		return ErrPoolClosed
	}

	w.recycleTime = time.Now()
	p.lock.Lock()
	defer p.lock.Unlock()

	p.workers = append(p.workers, w)
	p.cond.Signal()

	return nil
}

func (p *Pool) gc() {
	heartbeat := time.NewTicker(p.expiryDuration)
	defer heartbeat.Stop()

	for range heartbeat.C {
		if CLOSED == atomic.LoadInt32(&p.closed) {
			break
		}

		currentTime := time.Now()
		p.lock.Lock()
		idleWorkers := p.workers
		index := -1
		for i, w := range idleWorkers {
			if currentTime.Sub(w.recycleTime) <= p.expiryDuration {
				break
			}
			index = i
			w.task <- nil
			idleWorkers[i] = nil
		}
		if index > -1 {
			if index >= len(idleWorkers)-1 {
				p.workers = idleWorkers[:0]
			} else {
				p.workers = idleWorkers[index+1:]
			}
		}
		p.lock.Unlock()
	}
}
