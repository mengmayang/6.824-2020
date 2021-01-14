package main

import (
	"fmt"
	"time"
)

//

func main() {
	c := make(chan bool)
	go func() {
		time.Sleep(1 * time.Second)
		<-c //接收
	}()
	start := time.Now()
	c <- true //发送阻塞直到有接收
	fmt.Printf("send took %v\n", time.Since(start))
}