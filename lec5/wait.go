package main

//该程序使用channel，达到了与waitGroup一样的效果

func main() {
	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func(x int) {
			sendRPC(x)
			done <- true
		}(i)
	}
	for i := 0; i < 5; i++ {
		<- done
	}
}

func sendRPC(i int) {
	println(i)
}