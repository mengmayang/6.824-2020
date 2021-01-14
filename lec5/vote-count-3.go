package main

import (
	"sync"
	"time"
)

func main() {
	count := 0
	finished := 0
	var mu sync.Mutex

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
		}()
	}
	for {
		// vote-count-2 中for 这段代码虽然是正确的，却会让cpu的使用率达到100%，降低程序其他部分的效率
		mu.Lock()
		if count >= 5 || finished == 10{
			break
		}
		mu.Unlock()
		time.Sleep(5 * time.Millisecond) //一个简单的办法，解决cpu一直出于繁忙的问题
	}

	if count >=5 {
		println("received 5+ votes!")
	} else {
		println("lost")
	}
	mu.Unlock()
}

func requestVote() bool {
	return true
}