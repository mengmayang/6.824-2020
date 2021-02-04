package raft

import "log"

// Debugging
//const Debug = 1

var Debug int

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}


