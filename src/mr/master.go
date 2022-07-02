package mr

import (
	"log"
	"strconv"
	"sync"
	"time"
)
import "net"
import "os"
import "net/rpc"
import "net/http"

// global
var maptasks chan string // chan for map task
var reducetasks chan int // chan for reduce task

type Master struct {
	// Your definitions here.
	//用一个 map[string]int 类型的数据来存放 input 文件。
	//其中，map的 key 就代表 inuput 文件名，
	//int类型的 value 代表目前这个 input 文件的状态。
	//同样，我们使用一个 map[int]int 的数据来存放 reduce 任务。
	//int 类型的 key 代表 reduce 任务的编号， int 类型的 value 代表目前这个编号的 reduce 任务的状态。
	//另外用 int 类型的数据存放 reduce 任务数 和当前 Map 任务编号。用 bool 类型的数据存放是否完成整个 Map 和 Reduce。
	AllFilesName     map[string]int //splited files
	MapTaskNumCount  int            //curr map task number
	NReduce          int            //n reduce task
	InterFIlename    [][]string     // store location of intermediate files
	MapFinshed       bool
	ReduceTaskStatus map[int]int //about reduce task's status
	ReduceFinished   bool
	RWLock           *sync.RWMutex
}

// tasks' status
// UnAllocated    ---->     UnAllocated to any worker
// Allocated      ---->     be allocated to a worker
// Finished       ---->     worker finish the map task
const (
	UnAllocated = iota
	Allocated
	Finished
)

//generateTask : create tasks
func (m *Master) generateTask() {
	for k, v := range m.AllFilesName {
		if v == UnAllocated {
			maptasks <- k //add task to channel
		}
	}
	ok := false
	for !ok {
		ok = checkAllMapTask(m)
	}
	m.MapFinshed = true
	for k, v := range m.ReduceTaskStatus {
		if v == UnAllocated {
			reducetasks <- k
		}
	}
	ok = false
	for !ok {
		ok = checkAllReduceTask(m)
	}
	m.ReduceFinished = true
}

//将生成的任务，写入到相应的channel中去。
//worker 在向 master 请求任务的时候， master 从channel中获取到任务，发送给 worker ， worker执行。代码中还会一直监视任务的状态，来判断是否任务都完成了。
func checkAllMapTask(m *Master) bool {
	m.RWLock.RLock()
	defer m.RWLock.RUnlock()
	for _, v := range m.AllFilesName {
		if v != Finished {
			return false
		}
	}
	return true
}

func checkAllReduceTask(m *Master) bool {
	m.RWLock.RLock()
	defer m.RWLock.RUnlock()
	for _, v := range m.ReduceTaskStatus {
		if v != Finished {
			return false
		}
	}
	return true
}

// MyCallHandler func
// Your code here -- RPC handlers for the worker to call.
func (m *Master) MyCallHandler(args *MyArgs, reply *MyReply) error {
	msgType := args.MessageType
	// worker 发送的消息的类型
	switch msgType {
	case MsgForTask:
		select {
		case filename := <-maptasks:
			// allocate map task
			reply.Filename = filename
			reply.MapNumAllocated = m.MapTaskNumCount
			reply.NReduce = m.NReduce
			reply.TaskType = "map"

			m.RWLock.Lock()
			m.AllFilesName[filename] = Allocated
			m.MapTaskNumCount++
			m.RWLock.Unlock()
			go m.timerForWorker("map", filename)
			return nil

		case reduceNum := <-reducetasks:
			// allocate reduce task
			reply.TaskType = "reduce"
			reply.ReduceFileList = m.InterFIlename[reduceNum]
			reply.NReduce = m.NReduce
			reply.ReduceNumAllocated = reduceNum

			m.RWLock.Lock()
			m.ReduceTaskStatus[reduceNum] = Allocated
			m.RWLock.Unlock()
			go m.timerForWorker("reduce", strconv.Itoa(reduceNum))
			return nil
		}
	case MsgForFinishMap:
		// finish a map task
		m.RWLock.Lock()
		defer m.RWLock.Unlock()
		m.AllFilesName[args.MessageCnt] = Finished // set status as finish
	case MsgForFinishReduce:
		// finish a reduce task
		index, _ := strconv.Atoi(args.MessageCnt)
		m.RWLock.Lock()
		defer m.RWLock.Unlock()
		m.ReduceTaskStatus[index] = Finished // set status as finish
	}

	return nil
}

// MyInnerFileHandler : intermediate files' handler
func (m *Master) MyInnerFileHandler(args *MyIntermediateFile, reply *MyReply) error {
	nReduceNUm := args.NReduceType
	filename := args.MessageCnt
	//通过读取参数中的 NReduceType 字段，获取该文件应该由哪个编号的 reduce 任务处理，存放在相应的位置。
	// store themm
	m.InterFIlename[nReduceNUm] = append(m.InterFIlename[nReduceNUm], filename)
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
	// init channels
	maptasks = make(chan string, 5)
	reducetasks = make(chan int, 5)

	rpc.Register(m)
	rpc.HandleHTTP()

	// parallel run generateTask()
	go m.generateTask()
	//将 master 完成了 RPC 的注册，并完成了任务生成进程的运行。

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
	ret = m.ReduceFinished
	return ret
}

//
// create a Master.
// main/mrmaster.go calls this function.
// nReduce is the number of reduce tasks to use.
//
func MakeMaster(files []string, nReduce int) *Master {
	m := Master{}

	// Your code here.
	m.InterFIlename = make([][]string, m.NReduce)
	m.AllFilesName = make(map[string]int)
	m.MapTaskNumCount = 0
	m.NReduce = nReduce
	m.MapFinshed = false
	m.ReduceFinished = false
	m.ReduceTaskStatus = make(map[int]int)
	for _, v := range files {
		m.AllFilesName[v] = UnAllocated
	}
	for i := 0; i < nReduce; i++ {
		m.ReduceTaskStatus[i] = UnAllocated
	}
	//完善了一个基本的 Master 的结构。
	m.server()
	return &m
}

//按照我们 MyCallHandler 里的设计，我们在每分配一个任务后，即开始一个计时进程， 按照要求，以10秒为界限，监视任务是否被完成：
// TimerForWorker : monitor the worker
func (m *Master) timerForWorker(taskType, identify string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if taskType == "map" {
				m.RWLock.Lock()
				m.AllFilesName[identify] = UnAllocated
				m.RWLock.Unlock()
				// 重新让任务加入channel
				maptasks <- identify
			} else if taskType == "reduce" {
				index, _ := strconv.Atoi(identify)
				m.RWLock.Lock()
				m.ReduceTaskStatus[index] = UnAllocated
				m.RWLock.Unlock()
				// 重新将任务加入channel
				reducetasks <- index
			}
			return
		default:
			if taskType == "map" {
				m.RWLock.RLock()
				if m.AllFilesName[identify] == Finished {
					m.RWLock.RUnlock()
					return
				} else {
					m.RWLock.RUnlock()
				}
			} else if taskType == "reduce" {
				index, _ := strconv.Atoi(identify)
				m.RWLock.RLock()
				if m.ReduceTaskStatus[index] == Finished {
					m.RWLock.RUnlock()
					return
				} else {
					m.RWLock.RUnlock()
				}
			}
		}
	}
}
