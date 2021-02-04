package raft

import (
	"math/rand"
	"time"
)

const (
	ElectionTimeout  = time.Millisecond * 300
	HeartBeatTimeout = time.Millisecond * 150
	ApplyLogInterval = time.Millisecond * 100
	//MaxLockTime    = time.Millisecond * 10 // debug
	MaxLockTime      = 10 // debug
)

func (rf *Raft) resetElectionTimer() {
	rf.electionTimer.Stop()
	r := time.Duration(rand.Int63()) % ElectionTimeout
	rf.electionTimer.Reset(ElectionTimeout + r)
}

func (rf *Raft) resetHeartBeatTimers() {
	for peer := range rf.peers {
		rf.appendEntriesTimers[peer].Stop()
		rf.appendEntriesTimers[peer].Reset(0)
	}
}

func (rf *Raft) resetHeartBeatTimer(peer int) {
	rf.appendEntriesTimers[peer].Stop()
	rf.appendEntriesTimers[peer].Reset(HeartBeatTimeout)
}


