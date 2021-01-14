package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	alice := 10000
	bob := 10000
	var mu sync.Mutex
	total := alice + bob

	go func() {
		for i := 0; i < 1000; i++ {
			mu.Lock()
			bob = bob - 1
			alice = alice + 1
			mu.Unlock()
		}
	}()
	go func() {
		for i := 0; i < 1000; i++ {
			mu.Lock()
			alice = alice - 1
			bob = bob + 1
			mu.Unlock()
		}
	}()

	start := time.Now()
	for time.Since(start) < 1*time.Second {
		mu.Lock()
		if alice + bob != total {
			fmt.Printf("alice = %v, bob = %v, alice+bob = %v\n", alice, bob, alice + bob)
		}
		mu.Unlock()
	}
}