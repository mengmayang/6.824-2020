package main

import (
	"sync"
)

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	count := 0
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			mu.Lock()
			defer mu.Unlock()
			count = count + 1
			wg.Done()
		}()
	}

	wg.Wait()
	mu.Lock()
	println(count)
	mu.Unlock()
}
