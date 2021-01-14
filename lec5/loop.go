package main

import "sync"

func main()  {
	var wg sync.WaitGroup
	for i:=0; i < 5; i++ {
		// 并发：并行发送RPC
		// lab2中：如果candidate需要候选人投票，想同事从所有followers获取投票，而不是一个接着一个
		// 因为RPC是一种阻塞式操作，会耗一些时
		// 类似地，leader可能要给所有follower发送追加条目
		wg.Add(1)
		go func(x int) {
			sendRPC(x)
			wg.Done()
		}(i)// i是goroutine接收的外部作用域中的变量
		//将外部的i传递给匿名函数，并且可以将这个函数内部的变量名x重新命名
	}
	wg.Wait()
}

func sendRPC(i int) {
	println(i)
}
