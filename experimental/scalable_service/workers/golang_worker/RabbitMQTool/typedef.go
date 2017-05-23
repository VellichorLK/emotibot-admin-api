package rabbitmqtool

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/streadway/amqp"
)

type RabbitmqConnection struct {
	Host  string
	Port  int
	Conn  *amqp.Connection
	Lock  Mutex
	Count uint32
}

const mutexLocked = 1 << iota

type Mutex struct {
	sync.Mutex
}

func (m *Mutex) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(unsafe.Pointer(&m.Mutex)), 0, mutexLocked)
}
