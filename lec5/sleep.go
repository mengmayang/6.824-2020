package main

import "time"

//定时任务
func main() {
	time.Sleep(1 * time.Second)
	println("started")
	go periodic()
	time.Sleep(5 * time.Second)
}

func periodic() {
	for {
		println("tick")
		time.Sleep(1 * time.Second)
	}
}
