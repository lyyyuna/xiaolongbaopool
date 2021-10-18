package xiaolongbaopool_test

import (
	"sync"
	"testing"

	"github.com/lyyyuna/xiaolongbaopool"
)

const (
	RunTimes      = 1000000
	benchParam    = 10
	benchPoolSize = 200000
)

func BenchmarkGoroutine(b *testing.B) {
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			go func() {
				demo1()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkPool(b *testing.B) {
	var wg sync.WaitGroup
	p, _ := xiaolongbaopool.NewPool(benchPoolSize)
	defer p.Close()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(RunTimes)
		for j := 0; j < RunTimes; j++ {
			p.Submit(func() {
				demo1()
				wg.Done()
			})
		}
		wg.Wait()
	}
	b.StopTimer()
}
