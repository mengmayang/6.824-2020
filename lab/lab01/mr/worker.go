package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
)

//
// use ihash(key) % NReduce to choose the reduce
// Task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

type worker struct {
	//master需要知道task对应的是哪个worker处理的
	id       int
	mapf     func(string) []KeyValue
	reducef  func(string, []string) string
}

func (w *worker) register() {
	args := RegisterArgs{}
	reply := RegisterReply{}
	if ok := call("Master.RegisterWorker", &args, &reply); !ok {
		log.Fatal("worker register fail")
	}
	w.id = reply.WorkerId
}

func (w *worker) run() {
	for {
		args := TaskArgs{}
		args.WorkerId = w.id
		reply := TaskReply{}

		if ok := call("Master.AskTask", &args, &reply); !ok {
			DPrintf("worker get task fail, exit")
			os.Exit(1)
		}
		DPrintf("worker get task:%+v", reply.Task)
		task := *reply.Task
		//fmt.Printf("worker NReduce is %d\n:", task.NReduce)
		if !task.Alive {
			DPrintf("worker get task not alive, exit")
			return
		}
		if task.Phase == MapPhase {
			w.doMap(&task, w.mapf)
		} else if task.Phase == ReducePhase {
			w.doReduce(&task, w.reducef)
		}else {
			fmt.Println("task phase err!\n")
		}
	}
}

// worker:filename--> contents --> intermediate --> [] KeyValue{}

func Worker(mapf func(string) []KeyValue, reducef func(string, []string) string) () {
	// Your worker implementation here.
	// 1.获取filename
	// to do
	w := worker{}
	w.mapf = mapf
	w.reducef = reducef
	w.register()
	w.run()
	// uncomment to send the Example RPC to the master.
	//CallExample()
}

func (w *worker) doMap(task *Task, mapf func(string) []KeyValue) {
	filename := task.Filename
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal("cannnot open %v", filename)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("cannot read %v", filename)
	}
	file.Close()
	kva := mapf(string(content))
	reduces := make([][]KeyValue, task.NReduce)
	for _, kv := range kva {
		reduceId := ihash(kv.Key) % task.NReduce
		reduces[reduceId] = append(reduces[reduceId], kv)
	}

	for reduceId, arr := range reduces {
		if ok := task.GetTmpFileName(task.Seq, reduceId); !ok {
			DPrintf("get map-reduce tmp file fail\n")
		}
		f, err := os.Create(task.tmpFile)
		if err != nil {
			fmt.Println("open map-reduce tmp file fail!\n")
			w.reportTask(task, false, err)
			return
		}
		enc := json.NewEncoder(f)
		for _, kv := range arr {
			if err := enc.Encode(&kv); err != nil {
				fmt.Println("write map-reduce tmp file fail!\n")
				w.reportTask(task, false, err)
				return
			}
		}
		if err := f.Close(); err != nil {
			fmt.Println("close map-reduce tmp file fail!\n")
			w.reportTask(task, false, err)
		}
	}
	w.reportTask(task, true, nil)
}

func (w *worker) doReduce(task *Task, reducef func(string, []string) string) () {
	kva := []KeyValue{}
	for mapid := 0; mapid < task.NMaps; mapid++ {
		if ok := task.GetTmpFileName(mapid, task.Seq); !ok {
			DPrintf("get map-reduce tmp file fail\n")
		}
		f, err := os.Open(task.tmpFile)
		if err != nil {
			w.reportTask(task, false, err)
			return
		}
		dec := json.NewDecoder(f)
		for {
			var kv KeyValue
			if err := dec.Decode(&kv); err != nil {
				break
			}
			kva = append(kva, kv)
		}
	}
	sort.Sort(ByKey(kva))

	oname := "mr-out-"
	oname += strconv.Itoa(task.Seq)
	fmt.Printf("output file in doReduce is %s", oname)
	ofile, _ := os.Create(oname)

	i := 0
	for i < len(kva) {
		j := i + 1
		for j < len(kva) && kva[j].Key == kva[i].Key {
			j ++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, kva[k].Value)
		}
		// to do
		output := reducef(kva[i].Key, values)
		//写入结果文件
		if _, err := fmt.Fprintf(ofile, "%v %v\n", kva[i].Key, output); err != nil {
			panic(err)
		}
		i = j
	}
	ofile.Close()
	w.reportTask(task, true, nil)
}

func (w *worker) reportTask(task *Task, done bool, err error) () {
	//告诉master该task的状态
	if err != nil {
		log.Printf("%v", err)
	}
	args := ReportArgs{}
	args.Done = done
	args.WorkerId = w.id
	args.Seq = task.Seq
	args.Phase = task.Phase

	reply := ReportReply{}
	if ok := call("Master.ReportTask", &args, &reply); !ok {
		DPrintf("worker report task fail, exit")
		os.Exit(1)
	}

}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//

func CallExample() {
	args := ExampleArgs{}
	args.X = 99

	reply := ExampleReply{}

	call("Master.Example", &args, &reply)

	fmt.Printf("reply.Y %v\n", reply.Y)
}

func call(rpcname string, args interface{}, reply interface{}) bool {
	//这边的interface{}是什么类型？
	//go 语言规定，如果希望传递任意类型的变参，变参类型应该制定为空接口类型
	//空接口可以指向任何数据对象，所以可以使用interface{}定义任意类型变量，同时interface{}也是类型安全的
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := masterSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}