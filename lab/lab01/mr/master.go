package mr

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
	"fmt"
)

type Master struct {
	// Your definitions here.
	mu          sync.Mutex
	files       []string     //输入文件集
	nReduce     int          //定义使用worker的数量,并且输入的files会被划分成几个task给worker来处理
	taskCh      chan Task    //分发task的channel
	taskStats   []TaskStat   //记录每个task的状态
	taskPhase   TaskPhase
	done        bool
	workerSeq   int
}

// Your code here -- RPC handlers for the worker to call.
func (m *Master) initMapTask() {
	m.taskPhase = MapPhase
	m.taskStats = make([]TaskStat, len(m.files))
}

func (m *Master) initReduceTask() {
	DPrintf("init ReduceTask")
	m.taskPhase = ReducePhase
	m.taskStats = make([]TaskStat, m.nReduce)
}

func (m *Master) tickSchedule() {
	for !m.Done() {
		go m.schedule()
		time.Sleep(ScheduleInterval)
	}
}

func (m *Master) schedule() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.done {
		return
	}

	allFinish := true
	//遍历所有task的状态，启动的、没成功的放入taskCh缓冲通道
	for i, ts := range m.taskStats {
		if ts.Status == TaskStatusReady{
			allFinish = false
			m.taskCh <- m.getTask(i)
			m.taskStats[i].Status = TaskStatusQueue
		} else if ts.Status == TaskStatusQueue {
			allFinish = false
		} else if ts.Status == TaskStatusRunning {
			allFinish = false
			if time.Now().Sub(ts.StartTime) > MaxTaskRunTime {
				//超过运行的最大时长限制
				m.taskStats[i].Status = TaskStatusQueue
				m.taskCh <- m.getTask(i)
			}
		} else if ts.Status == TaskStatusErr {
			allFinish = false
			m.taskStats[i].Status = TaskStatusQueue
			m.taskCh <- m.getTask(i)
		}
	}
	if allFinish {
		if m.taskPhase == MapPhase {
			m.initReduceTask()
		} else {
			m.done = true
		}
	}
}

func (m *Master) getTask(taskSeq int) Task {
	task := Task{}
	task.Filename = ""
	task.NReduce = m.nReduce
	task.NMaps = len(m.files)
	task.Seq = taskSeq
	task.Phase = m.taskPhase
	task.Alive = true

	DPrintf("m:%+v, taskseq:%d, lenfiles:%d, lentasks:%d", m, taskSeq, len(m.files), len(m.taskStats))
	if task.Phase == MapPhase {
		task.Filename = m.files[taskSeq]
	}
	return task
}

func (m *Master) AskTask(args *TaskArgs, reply *TaskReply) error {
	//master负责通过channel来分发task，task需要初始化，更新状态
	task := <- m.taskCh
	reply.Task = &task

	if task.Alive {
		m.mu.Lock()
		defer m.mu.Unlock()

		if task.Phase != m.taskPhase {
			panic("req Task phase neq") //抛出异常
		}
		m.taskStats[task.Seq].Status = TaskStatusRunning
		m.taskStats[task.Seq].StartTime = time.Now()
		m.taskStats[task.Seq].WorkerId = args.WorkerId
	}

	DPrintf("master response the request from worker, args:%+v, reply:%+v", args, reply)
	return nil
}

func (m *Master) RegisterWorker(args *RegisterArgs, reply *RegisterReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workerSeq += 1
	reply.WorkerId = m.workerSeq
	return nil
}

//
// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
//

func (m *Master) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

//
// start a thread that listens for RPCs from worker.go
//
func (m *Master) server() {
	rpc.Register(m)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := masterSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

//
// main/mrmaster.go calls Done() periodically to find out
// if the entire job has finished.
//
func (m *Master) Done() bool {
	ret := false
	// Your code here.
	m.mu.Lock()
	defer m.mu.Unlock()
	ret = m.done
	return ret
}

func (m *Master) ReportTask (args *ReportArgs, reply *ReportReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	DPrintf("get report task: %+v, taskPhase: %+v", args, m.taskPhase)

	if m.taskPhase != args.Phase || args.WorkerId != m.taskStats[args.Seq].WorkerId {
		return nil
	}

	if args.Done {
		m.taskStats[args.Seq].Status = TaskStatusFinish
	} else {
		m.taskStats[args.Seq].Status = TaskStatusErr
	}

	go m.schedule()
	return nil
}


//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//

func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here
	// 根据files和nReduce来分任务task
	m.mu = sync.Mutex{} //to explain
	m.nReduce = nReduce
	fmt.Printf("master nReduce is %d\n:", m.nReduce)
	m.files = files

	// 创建一个有缓冲通道, 大小为reduce数量和map数量中大的那个
	if nReduce > len(files){
		m.taskCh = make(chan Task, nReduce)
	}else {
		m.taskCh = make(chan Task, len(files))
	}

	m.initMapTask() //to finished
	go m.tickSchedule() //启动一个go协程去触发master调度
	m.server()
	DPrintf("master init")
	return &m
}


