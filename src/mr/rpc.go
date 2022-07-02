package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"fmt"
	"os"
)
import "strconv"

//
// example to show how to declare the arguments
// and reply for an RPC.
//
// Add your RPC definitions here.
const (
	MsgForTask         = iota // ask a task
	MsgForInterFileLoc        // send intermediate files' location to master
	MsgForFinishMap           // finish a map task
	MsgForFinishReduce        // finish a reduce task
)

type MyArgs struct {
	MessageType int
	MessageCnt  string
}

//MyIntermediateFile 专用于发送中间文件的位置到master。
//即消息类型为MsgForInterFileLoc的使用 MyIntermediateFile。其他消息类型使用 MyArgs。
// send intermediate files' filename to master
type MyIntermediateFile struct {
	MessageType int
	MessageCnt  string
	NReduceType int
}

//所有 RPC 请求的 reply 均使用该类型的 reply。
//struct 中包括： 被分配的任务的类型，map 任务的话，FIlename字段装载input文件名，
//reduce 任务的话， ReduceFileList 字段装载文件名的list。
//MapNumAllocated 代表 map 任务被分配的任务的编号。
//ReduceNumAllocated 代表 reduce 任务被分配的任务的编号。NReduce 字段代表总 reduce 任务数。
type MyReply struct {
	Filename           string // get a filename
	MapNumAllocated    int
	NReduce            int
	ReduceNumAllocated int
	TaskType           string   // refer a task type : "map" or "reduce"
	ReduceFileList     []string // File list about
}

type ExampleArgs struct {
	X int
}

type ExampleReply struct {
	Y int
}

// SendInterFiles : send intermediate files' location (filenames here) to master
func SendInterFiles(msgType int, msgCnt string, nReduceType int) MyReply {
	args := MyIntermediateFile{}
	args.MessageType = msgType
	args.MessageCnt = msgCnt
	args.NReduceType = nReduceType

	repley := MyReply{}

	res := call("Master.MyInnerFileHandler", &args, &repley)
	if !res {
		fmt.Println("error sending intermediate files' location")
	}
	return repley
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the master.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func masterSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
