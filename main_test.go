package main_test

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	wg:= &sync.WaitGroup{}
	for i:= 0 ; i < 10 ; i++ {
		go func() {
			wg.Add(1)
			for {
				time.Sleep(3 * time.Second)
				wg.Done()
			}
		}()
	}

	totalCpu := runtime.NumCPU()
	totalThread := runtime.GOMAXPROCS(-1)
	totalGoroutine := runtime.NumGoroutine()
	println("Total CPU: ", totalCpu)
	println("Total Thread: ", totalThread)
	println("Total Goroutine: ", totalGoroutine)

	wg.Wait()
}
