package xiaolongbaopool_test

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/lyyyuna/xiaolongbaopool"
)

const (
	_   = 1 << (10 * iota)
	KiB // 1024
	MiB // 1048576
	GiB // 1073741824
	TiB // 1099511627776             (超过了int32的范围)
	PiB // 1125899906842624
	EiB // 1152921504606846976
	ZiB // 1180591620717411303424    (超过了int64的范围)
	YiB // 1208925819614629174706176
)

const (
	Param    = 100
	PoolSize = 1000
	TestSize = 10000
	n        = 100000
)

var curMem uint64

func demo1() {
	time.Sleep(time.Duration(10) * time.Millisecond)
}

func demo2(i int) {
	time.Sleep(time.Duration(i) * time.Millisecond)
}

func TestPoolWaitToGetWorker(t *testing.T) {
	var wg sync.WaitGroup
	p, err := xiaolongbaopool.NewPool(PoolSize)
	if err != nil {
		t.Fatalf("new pool failed: %v", err)
	}
	defer p.Close()

	for i := 0; i < n; i++ {
		wg.Add(1)
		p.Submit(func() {
			demo1()
			wg.Done()
		})
	}
	wg.Wait()

	t.Logf("running workers number: %v", p.Running())

	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	curMem = mem.TotalAlloc/MiB - curMem
	t.Logf("memory usage: %v MB", curMem)
}
