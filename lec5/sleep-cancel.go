package main

import (
	"sync"
	"time"
)

var done bool
var mu sync.Mutex

// 定时任务，通过使用一个共享变量来控制线程，巨顶这个goroutine是否结束
func main() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodic()
	time.Sleep(5 * time.Second)
	mu.Lock()
	done = true
	mu.Unlock()
	println("canceled")
	time.Sleep(3 * time.Second)
}

func periodic() {
	for {
		println("tick")
		time.Sleep(1 * time.Second)
		mu.Lock()
		if done {
			mu.Unlock()
			return
		}
		mu.Unlock()
	}
}
