package main

import "sync"

func main()  {
	var a string
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		//这个叫闭包closure，这个匿名函数可以引用这个封闭作用域内的任何变量
		a = "hello world"
		wg.Done()
	}()
	wg.Wait()
	println(a)

}
