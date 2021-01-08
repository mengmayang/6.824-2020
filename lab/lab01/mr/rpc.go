package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"log"
	"os"
	"time"
)
import "strconv"

type KeyValue struct {
	Key string
	Value string
}

//ByKey 需要实现sort.Interface
type ByKey []KeyValue

func (a ByKey) Len() int { return len(a) }
func (a ByKey) Less(i, j int) bool {return a[i].Key < a[j].Key}
func (a ByKey) Swap(i, j int) {a[i], a[j] = a[j], a[i]}

type TaskPhase int
type TaskStatus int

const (
	MapPhase TaskPhase = 0
	ReducePhase TaskPhase = 1
)

const (
	TaskStatusReady TaskStatus   = 0
	TaskStatusQueue TaskStatus   = 1
	TaskStatusRunning TaskStatus = 2
	TaskStatusFinish TaskStatus  = 3
	TaskStatusErr TaskStatus     = 4
)

const (
	ScheduleInterval = time.Millisecond * 500
	MaxTaskRunTime = time.Second * 5
)

const DEBUG = true

type TaskArgs struct {
	WorkerId int
}

type TaskReply struct {
	Task *Task
}

type RegisterArgs struct {
}

type RegisterReply struct {
	WorkerId int
}

type ReportArgs struct {
	Done      bool
	Seq       int
	Phase     TaskPhase
	WorkerId  int
}

type ReportReply struct {
}

type Task struct {
	//task分为map任务和reduce任务
	//一个map任务处理一个file
	Filename string //输入的file
	Alive    bool
	Phase    TaskPhase
	NReduce  int
	Seq      int //task的序号，也是map任务的序号, 或者reduce任务的序号
	NMaps    int //一个file对应一个map，执行一次doMap，n个files就对应nMaps
	tmpFile  string
}

type TaskStat struct {
	Status    TaskStatus
	WorkerId  int
	StartTime time.Time
}

func (task *Task) GetTmpFileName (mapid int, reduceId int) bool {
	task.tmpFile = fmt.Sprintf("mr-%d-%d", mapid, reduceId)
	return true
}

//
// example to show how to declare the arguments
// and reply for an RPC.
//

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// Add your RPC definitions here.


// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.

func masterSock() string {
	s := "824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}

func DPrintf(format string, v...interface{}) {
	if DEBUG {
		log.Printf(format+"\n", v...)
	}
}

