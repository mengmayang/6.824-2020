package main

//发送会被阻塞，只有有人准备去接收时才不会被阻塞，这个代码里并没有接收者，golang会报错
//fatal error: all goroutines are asleep - deadlock!
//如果你所有线程都在沉睡，那么Go就检测到这里发生了死锁
//channel并不能在单线程中被使用，没有意义，它的初衷就是为了去发送或者接收数据

func main() {
	c := make(chan bool)
	c <- true
	<-c
}
