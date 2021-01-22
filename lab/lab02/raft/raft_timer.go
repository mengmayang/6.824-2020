package raft

import (
	"math/rand"
	"time"
)

const (
	ElectionTimeout       = time.Millisecond * 2000
	HeartBeatInterval     = time.Millisecond * 150
	AppendEntriesInterval = time.Millisecond * 100
	ApplyLogInterval      = time.Millisecond * 100
	ElectionTimer         = 1
	HeartBeatTimer        = 2
	AppendEntriesTimer    = 3
)

type RaftTimer struct{
	timerType  int
	timer      *time.Timer
	interval   time.Duration
}

func (rt *RaftTimer) setTimer(timerType int) {
	rt.timerType = timerType
	if rt.timerType == ElectionTimer {
		rt.interval = ElectionTimeout
	}else if rt.timerType == HeartBeatTimer {
		rt.timerType = HeartBeatTimer
		rt.interval = HeartBeatInterval
	}else if rt.timerType == AppendEntriesTimer {
		rt.timerType = AppendEntriesTimer
		rt.interval = AppendEntriesInterval
	}
	r := time.Duration(rand.Int63()) % rt.interval
	rt.timer = time.NewTimer(rt.interval + r) // golang 定时器，定期向自身的C字段发送当时的时间
}

func (rt *RaftTimer) resetTimer() {
	rt.timer.Stop()
	r := time.Duration(rand.Int63()) % rt.interval
	//DPrintf("%v", rt.interval + r)
	rt.timer.Reset(rt.interval + r)
}