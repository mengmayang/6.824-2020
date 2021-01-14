package main

import "sync"

func main()  {
	count := 0
	finished := 0
	var mu sync.Mutex
	cond := sync.NewCond(&mu) //条件变量，当共享数据中的某个条件或属性变为true时，通过条件变量来协调

	for i := 0; i < 10; i++ {
		go func() {
			vote := requestVote()
			mu.Lock()
			defer mu.Unlock()
			if vote {
				count++
			}
			finished++
			cond.Broadcast()
		}()
	}
	mu.Lock()
	for count < 5 && finished != 10{
		cond.Wait()
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