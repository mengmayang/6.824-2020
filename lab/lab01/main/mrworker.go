package main

//
// start a worker process, which is implemented
// in ../mr/worker.go. typically there will be
// multiple worker processes, talking to one master.
//
// go run mrworker.go wc.so
//
// Please do not change this file.
//

import (
	"fmt"
	"lab01/mr"
	"lab01/wc"
	"os"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: mrworker\n")
		os.Exit(1)
	}

	// 由于windows平台对plugin不支持，所以把mapf, reducef := loadPlugin(os.Args[1])改成直接函数赋值
	var mapf func(string) []mr.KeyValue
	var reducef func(string, []string) string

	mapf = wc.Map
	reducef = wc.Reduce

	mr.Worker(mapf, reducef)
}