package handlers

import (
	"net/http"
	"sync/atomic"
)

var connPool ConnectionPool

//InitController init the controller for rabbitmq
func InitController(host string, port int) {
	rc := &RabbitmqConnection{Host: host, Port: port}
	rc.ConnectToRabbitMQ()
	connPool.Conns = append(connPool.Conns, rc)
	connPool.Counter++

}

//PushTask push the task into queue, task: task string, ms: timeout in millisecond unit. return result, content_type, status code
func PushTask(task string, taskQueue string, ms int) (string, string, int) {
	//currently we only use one conn, maybe create the more connections with some mechaism later later..
	conn := connPool.Conns[0]
	count := atomic.LoadUint32(&conn.Count)

	if count >= MAXCHANNEL {
		return "Too many channel!", "text/plain", http.StatusTooManyRequests
	}

	res, contentType, status := InitChannelAndSendTask(conn, task, taskQueue, ms)

	if status == http.StatusServiceUnavailable {
		go conn.Reconnect()
	}

	return res, contentType, status
}
