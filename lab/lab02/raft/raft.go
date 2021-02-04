package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// The peers argument is an array of network identifiers of the Raft peers (including this one), for use with RPC. The me argument is the index of this peer in the peers array.

// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// asks Raft to start the processing to append the command to the replicated log. Start() should return immediately, without waiting for the log appends to complete.

// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader

// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
// The service expects your implementation to send an ApplyMsg for each newly committed log entry to the applyCh channel argument to Make().
//

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)
import "sync/atomic"
import "lab02/labrpc"

// import "bytes"
// import "../labgob"

//
// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in Lab 3 you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh; at that point you can add fields to
// ApplyMsg, but set CommandValid to false for these other uses.
//

// 要发送给server的（tester/kv server）
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int
}

// Define a struct to hold information about each log entry
type Entry struct {
	Term   int
	Index  int
	Command  interface{}
}

//
// A Go object implementing a single Raft peer.
//
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	// Figure2
	// Updated on stable storage before responding to RPCs 在响应RPC前更新持久化的存储
	//
	// Persistent State
	currentTerm          int
	voteFor              int
	log                  []Entry

	// Volatile State
	commitIndex          int
	lastApplied          int

	// Volatile State on leaders
	nextIndex            []int
	matchIndex           []int

	// added by me
	state                string
	applyCh              chan ApplyMsg
	voteCountCh          chan bool
	electionTimer        *time.Timer
	appendEntriesTimers   []*time.Timer
	applyTimer           *time.Timer
	lockName             string
	lockStart            time.Time
	lockEnd              time.Time
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {

	var term int
	var isleader bool
	// Your code here (2A).
	term = rf.currentTerm
	if rf.state == "leader" {
		isleader = true
	} else {
		isleader = false
	}
	//DPrintf("GetState of %d, term=%d, isleader=%v", rf.me, term, isleader)
	return term, isleader
}

//
// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
//
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}


//
// restore previously persisted state.
//
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}


//
// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
//
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	//Client 给复制状态机发送了一个command
	rf.mu.Lock()
	index := len(rf.log)
	term := rf.currentTerm
	isLeader := true

	// Your code here (2B).
	//DPrintf("2B TEST: current server is %d, state is %s", rf.me, rf.state)
	if rf.state != "leader" {
		isLeader = false
	}else{
		if rf.log[0].Term == 0 && rf.log[0].Index == 0 && rf.log[0].Command == nil {
			rf.log[0].Term = rf.currentTerm
			rf.log[0].Command = command
		}else{
			index = len(rf.log)
			logEntry := Entry{
				Term:    rf.currentTerm,
				Index:   index,
				Command: command,
			}
			rf.log = append(rf.log, logEntry)
			rf.matchIndex[rf.me] = index
			DPrintf("2B TEST: %d's log is %v; index is %d, term is %d, logEntry is %v", rf.me, rf.log, index, term, logEntry)
		}
	}
	rf.resetHeartBeatTimers()
	rf.mu.Unlock()
	return index, term, isLeader
}

//
// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
//
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

func (rf *Raft) lock(m string) {
	rf.mu.Lock()
	rf.lockStart = time.Now().UTC()
	rf.lockName = m
}

func (rf *Raft) unlock(m string) {
	rf.lockEnd = time.Now().UTC()
	lockDuration := rf.lockEnd.Sub(rf.lockStart).Nanoseconds()/1e6
	//DPrintf("%d %s unlock, rf.lockStart is %v, rf.lockEnd is %v, duration is %v", rf.me, m, rf.lockStart, rf.lockEnd, lockDuration)
	rf.lockName = ""
	if rf.lockName != "" && lockDuration > MaxLockTime {
		//DPrintf("%d's %s lock too long!iskill:%v", rf.me, rf.lockName, rf.killed())
	}
	rf.mu.Unlock()
}


//
// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int
	CandidatedId int
	LastLogIndex int
	LastLogTerm  int
}

//
// example RequestVote RPC reply structure.
// field names must start with capital letters!
//
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int
	VoteGranted bool
}

//
// example RequestVote RPC handler.
//
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	//DPrintf("%d received RequestVote from %d", rf.me, args.CandidatedId)
	rf.mu.Lock()
	// Your code here (2A, 2B).
	rf.resetElectionTimer()
	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	lastLogIndex := len(rf.log)-1
	lastLogTerm := rf.log[lastLogIndex].Term

	if args.Term < rf.currentTerm {
		rf.mu.Unlock()
		return
	}

	if args.Term == rf.currentTerm {
		if rf.state == "leader" {
			rf.mu.Unlock()
			return
		}
		if rf.voteFor == args.CandidatedId {
			reply.VoteGranted = true
			rf.mu.Unlock()
			return
		}else if rf.voteFor != -1 {
			rf.mu.Unlock()
			return
		}
	}else{//if args.Term > rf.currentTerm {

	// Rules for Servers
	// All Servers
	// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower

		rf.currentTerm = args.Term
		//DPrintf("%d is transfor to follower", rf.me)
		rf.mu.Unlock()
		rf.changeState("follower")
		rf.mu.Lock()
	}
	if lastLogTerm > args.LastLogTerm  || (args.LastLogTerm == lastLogTerm && args.LastLogIndex < lastLogIndex) {
		rf.mu.Unlock()
		return
	}

	rf.currentTerm = args.Term
	rf.voteFor = args.CandidatedId
	reply.VoteGranted = true
	rf.mu.Unlock()
	rf.changeState("follower")
	return
}

func (rf *Raft) changeState(state string) {
	//DPrintf("%d is %s ,and changeState to %s", rf.me, rf.state, state)
	rf.lock("changeState")
	rf.state = state
	//if state == "candidate" && rf.state == "follower"{
	if rf.state == "candidate" { //leader 长时间收不到回复也会election timeout转化成candidate
		// 1.Increment current Term
		rf.currentTerm += 1
		// 2.Vote for self
		rf.voteFor = rf.me
		// 3.Reset election timer
		rf.resetElectionTimer()
		rf.unlock("changeState")
		rf.startElection()
		return
	}
	if rf.state == "leader" {
		//rf.electionTimer.timer.Stop()
		rf.resetHeartBeatTimers()
		rf.nextIndex = make([]int, len(rf.peers))
		rf.matchIndex = make([]int, len(rf.peers))
		// index of candidate's last log entry
		//lastLogIndex := len(rf.log) - 1 // initialized to leader last log index  (leader last log index = lastSnapshotIndex + len(rf.log) - 1) 这里先不考虑快照，假设只有log
		lastLogIndex := rf.log[len(rf.log)-1].Index
		for i := 0; i < len(rf.peers); i++ {
			if rf.nextIndex[i] != 0 {
				rf.nextIndex[i] = lastLogIndex + 1
			}
			rf.matchIndex[i] = 0 // initialized to 0
		}
		rf.resetElectionTimer()
	}
	if rf.state == "follower" {
		rf.voteFor = -1
		rf.resetElectionTimer()
	}
	rf.unlock("changeState")
}

func (rf *Raft) startElection() {
	rf.lock("startElection1")
	//DPrintf("%d is %s, and start election", rf.me, rf.state)
	if rf.state != "candidate" {
		rf.unlock("startElection2")
		return
	}
	// 0.Initailize voteCountCh
	rf.voteCountCh = make(chan bool, len(rf.peers)) //有缓存的通道
	//DPrintf("%d start Election and term is %d", rf.me, rf.currentTerm)

	// Sent Request Vote RPCs to all other servers
	lastLogIndex := len(rf.log) - 1
	lastLogTerm := rf.log[len(rf.log) - 1].Term
	requestVoteArgs := RequestVoteArgs{
		Term          : rf.currentTerm,
		CandidatedId  : rf.me,
		LastLogIndex  : lastLogIndex,
		LastLogTerm   : lastLogTerm,
	}

	chPeerCount := 1
	grantedCount := 1

	rf.unlock("startElection3")
	for peer := range(rf.peers) {
		// If votes received from majority of servers:become leader
		if peer == rf.me {
			continue
		}
		//DPrintf("%d request vote from %d", rf.me, peer)
		go func(ch chan bool, peer int) {
			requestVoteReply := RequestVoteReply{}
			rf.sendRequestVote(peer, &requestVoteArgs, &requestVoteReply)
			ch <- requestVoteReply.VoteGranted
			// Rules for Servers
			// All Servers
			// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
			rf.lock("startElection4")
			if requestVoteReply.Term > rf.currentTerm {
				rf.unlock("startElection5")
				rf.changeState("follower6")
				rf.currentTerm = requestVoteReply.Term
				//DPrintf("%s will change to follower", rf.state)
				rf.lock("startElection7")
			}
			rf.unlock("startElection8")
		}(rf.voteCountCh, peer)
	}

	for {
		//DPrintf("===========================================chPeerCount is %d", chPeerCount)
		r := <-rf.voteCountCh
		chPeerCount += 1
		//DPrintf("===========================================chPeerCount is %d, voteCount is %v", chPeerCount, r)
		if r == true {
			grantedCount += 1
			//DPrintf("===========================================%d received %d vote", rf.me, grantedCount)
		}
		if chPeerCount == len(rf.peers) || grantedCount > len(rf.peers)/2 || chPeerCount-grantedCount > len(rf.peers)/2 {
			break
		}
	}

	rf.lock("startElection9")
	//DPrintf("%d: grantedCount[%d]<=len(rf.peers)/2[%d], rf.peers=%d", rf.me, grantedCount, len(rf.peers)/2, len(rf.peers))
	if grantedCount <= len(rf.peers)/2 {
		rf.unlock("startElection10")
		return
	}
	//DPrintf("%d startElection, currentTerm is %d, args.term=%d, state is %s", rf.me, rf.currentTerm, requestVoteArgs.Term, rf.state)

	if rf.currentTerm == requestVoteArgs.Term && rf.state == "candidate" {
		rf.unlock("startElection11")
		//DPrintf("%d changeState to leader", rf.me)
		rf.changeState("leader")
		return
	}
	rf.lock("startElection12")
	if rf.state == "leader" {
		rf.resetHeartBeatTimers()
	}
	rf.unlock("startElection13")
	return
}

//
// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
//
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	//DPrintf("%d sendRequestVote to %d", rf.me, server)
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

//
// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
//
func Make(peers []*labrpc.ClientEnd, me int, persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{} //一个raft实例
	rf.peers = peers //一个raft实例包含的所有servers
	rf.persister = persister //存放这台机器的持久状态persistent state
	rf.me = me

	// Your initialization code here (2A, 2B, 2C).
	rf.currentTerm = 0 //initialized to 0 on first boot
	rf.state = "follower"
	rf.voteFor = -1 // null if none
	rf.log = make([]Entry, 1)
	rf.commitIndex = 0 //initialized to 0
	rf.lastApplied = 0 //initialized to 0
	r := time.Duration(rand.Int63()) % ElectionTimeout
	rf.electionTimer = time.NewTimer(ElectionTimeout + r)
	rf.appendEntriesTimers = make([]*time.Timer, len(rf.peers))
	for peer := range(rf.peers) {
		rf.appendEntriesTimers[peer] = time.NewTimer(ElectionTimeout)
	}
	rf.applyCh = applyCh

	// 选举定时器
	go func() {
		//DPrintf("选举定时器")
		for {
			//if rf.state != "leader" {
			//	<-rf.electionTimer.timer.C // 定时器
			//	//DPrintf("%d is %s, and change to candidate", rf.me, rf.state)
			//	//if rf.state == "follower" {
			//	rf.changeState("candidate")
			//	//}
			//	//rf.mu.Unlock()
			//} else {
			//	rf.electionTimer.timer.Stop()
			//}
			// 即使被选为leader，选举定时器也不能停止，因为如果一旦有peer down出现，并且达不到quorum 法定人数，则不允许有leader被选出
			<-rf.electionTimer.C // 定时器
			rf.changeState("candidate")
		}
	}()

	// 发送appendEntries定时器
	for peer := range(rf.peers) {
		if peer == rf.me {
			continue
		}
		go func(peer int) {
			for {
				<-rf.appendEntriesTimers[peer].C
				if rf.state == "leader" {
					DPrintf("%d is leader and term is %d, send appendEntries to %d", rf.me, rf.currentTerm, peer)
					rf.appendEntries2Peer(peer)
				}
			}
		}(peer)
	}

	go func() {
		for !rf.killed() {
			time.Sleep(2000 * time.Millisecond)
			if rf.lockName != "" {
				fmt.Printf("%d who has lock: %s; iskilled:%v; duration: %v; MaxLockTime is :%v; rf.loclkStart: %v; rf.lockEnd: %v\n", rf.me, rf.lockName, rf.killed(), rf.lockEnd.Sub(rf.lockStart).Nanoseconds()/1e6, MaxLockTime, rf.lockStart, rf.lockEnd)
			}
		}
	}()

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())
	return rf
}

func (rf *Raft) appendEntries2Peer(peer int) {
	rf.lock("appendEntries2Peer")
	//DPrintf("%d send append entries to %d", rf.me, peer)
	rf.resetHeartBeatTimer(peer)
	rf.resetElectionTimer()
	lastLogIndex := len(rf.log) - 1
	prevLogIndex := rf.nextIndex[peer] - 1
	if rf.nextIndex[peer] > lastLogIndex {
		// 没有需要发送的log
		prevLogIndex = lastLogIndex
	}
	if prevLogIndex < 0{
		prevLogIndex = 0
	}
	prevLogTerm := rf.log[prevLogIndex].Term
	//logs := append([]Entry{}, rf.log[rf.nextIndex[peer]:]...)
	logs := rf.log[rf.nextIndex[peer]:]

	appendEntriesArgs:= AppendEntriesArgs{
		Term          : rf.currentTerm,
		LeaderId      : rf.me,
		PrevLogIndex  : prevLogIndex,
		PrevLogTerm   : prevLogTerm,
		Entries       : logs,
		LeaderCommit  : rf.commitIndex,
	}
	DPrintf("appendEntries2Peer appendEntriesArgs is : %v, rf.nextIndex[peer] : %v", appendEntriesArgs, rf.nextIndex[peer])

	appendEntriesReply := AppendEntriesReply{}
	rf.unlock("appendEntries2Peer")
	//DPrintf("%d is leader and send appendEntries to %d", rf.me, peer)
	rf.sendAppendEntries(peer, &appendEntriesArgs, &appendEntriesReply)
	// Rules for Servers
	// All Servers
	// If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	rf.lock("appendEntries2Peer")
	if appendEntriesReply.Term > rf.currentTerm {
		//DPrintf("%d is leader and will change to follow", rf.me)
		rf.unlock("appendEntries2Peer")
		rf.changeState("follower")
		rf.lock("appendEntries2Peer")
		rf.currentTerm = appendEntriesReply.Term
	}
	rf.unlock("appendEntries2Peer")
}

func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []Entry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.lock("AppendEntries")
	rf.resetElectionTimer()
	//DPrintf("%d received %d append entries, args.appendEntriesArgs:%v", rf.me, args.LeaderId, args.Entries)

	// currentTerm, for leader to update itself
	reply.Term = rf.currentTerm
	//DPrintf("%d received appendEntries from %d, %d term is %d, leader term is %d", rf.me, args.LeaderId, rf.me, rf.currentTerm, args.Term)
	//rf.electionTimer.resetTimer()

	// 5.1. Reply false if term < currentTerm
	if args.Term < rf.currentTerm {
		reply.Success = false
		rf.unlock("AppendEntries")
		return
	}
	// Rules for Servers
	// All Servers
	// 5.1 If RPC request or response contains term T > currentTerm: set currentTerm = T, convert to follower
	rf.currentTerm = args.Term
	rf.unlock("AppendEntries")
	rf.changeState("follower")
	rf.lock("AppendEntries")

	if rf.log[0].Command == nil && len(rf.log) == 1{
		rf.log = args.Entries
	}

	// 5.3 Reply false if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	flag := 0
	for i := args.PrevLogIndex; i < len(rf.log); i++ {
		for j := 0; j < len(args.Entries); j++ {
			if args.Entries[j].Term == rf.log[i].Term {
				if i == j {
					flag = 1
				}
			}
		}
	}
	if flag != 1 {
		reply.Success = false
	}

	// 5.3 If an existing entry conflicts with a new one(same index but different terms), delete the existing entry and all that follow it
	conflict := -1
	cycle:
	for i := args.PrevLogIndex; i < len(rf.log); i++ {
		for j := 0; j < len(args.Entries); j++ {
			if rf.log[i].Term == args.Entries[j].Term && rf.log[i].Index == args.Entries[j].Index && rf.log[i].Command != args.Entries[j].Command {
				conflict = i
				break cycle
			}
		}
	}
	DPrintf("args.PrevLogIndex : %d, args.Entries: %v, rf.log: %v", args.PrevLogIndex, args.Entries, rf.log)
	if conflict > -1 {
		for i := conflict; i < len(rf.log); i++ {
			rf.log[i] = Entry{}
		}
	}

	// Append any new entries not already in the log
	match := -1
	for i := 0; i < len(args.Entries); i++ {
		for j := args.PrevLogIndex; j < len(rf.log); j++ {
			if rf.log[j].Term == args.Entries[i].Term && rf.log[j].Index == args.Entries[i].Index && rf.log[j].Command == args.Entries[i].Command {
				match = i + 1
			}
		}
	}

	if match != -1 {
		rf.log = append(rf.log, args.Entries[match:]...)
	}else {
		rf.log = append(rf.log, args.Entries[0:]...)
	}

	// If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rf.commitIndex {
		if args.LeaderCommit > rf.log[len(rf.log)-1].Index {
			rf.commitIndex = rf.log[len(rf.log)-1].Index
		}else{
			rf.commitIndex = args.LeaderCommit
		}
	}

	rf.commit()

	DPrintf("%d 's log is :%v", rf.me, rf.log)
	rf.unlock("AppendEntries")
}

func (rf *Raft) commit() {

	// 要发送给server的（tester/kv server）
	//type ApplyMsg struct {
	//	CommandValid bool
	//	Command      interface{}
	//	CommandIndex int
	for i := rf.lastApplied; i <= rf.commitIndex; i ++ {
		msg := ApplyMsg {
			CommandValid: true,
			Command: rf.log[i].Command,
			CommandIndex: i,
		}
		rf.applyCh <- msg
	}
}
