package handlers

import (
	"net/http"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/streadway/amqp"
)

type ModuleHandleFunc map[string]func(http.ResponseWriter, *http.Request)

type YamlCfg struct {
	MethodData map[string]FuncUrl `yaml:"paths"`
}

type FuncUrl map[string]interface{}

//FuncSupportMethod supported url path and its relative data
type FuncSupportMethod map[string]*SupportData

//SupportData Method: supported method and its return data type, Queue: its task queue name
type SupportData struct {
	Method map[string]string
	Queue  string
}

type RabbitmqConnection struct {
	Host  string
	Port  int
	Conn  *amqp.Connection
	Lock  Mutex
	Count uint32
}

type ConnectionPool struct {
	Conns   []*RabbitmqConnection
	Counter uint32
}

const MAXCHANNEL = 30000
const MAXCONNECTIONS = 2000

const mutexLocked = 1 << iota

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked)
}

type TaskBlock struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   string `json:"body"`
	Query  string `json:"query"`
}
