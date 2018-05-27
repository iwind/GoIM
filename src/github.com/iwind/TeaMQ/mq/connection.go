package mq

import (
	"github.com/iwind/TeaMQ/nets"
	"sync"
	"strings"
	"encoding/json"
)

type Connection struct {
	userId   int64
	queues   map[string]int
	client   *nets.Client
	isWorker bool

	mutex *sync.Mutex
}

func NewConnection(client *nets.Client) *Connection {
	var connection = &Connection{
		queues: map[string]int{},
		client: client,
		mutex:  &sync.Mutex{},
	}
	return connection
}

func (connection *Connection) Id() int {
	return connection.client.Id()
}

func (connection *Connection) IsAuthenticated() bool {
	return connection.userId > 0
}

func (connection *Connection) setUserId(userId int64) {
	connection.userId = userId
}

func (connection *Connection) SubscribeQueue(queue string) {
	queue = strings.TrimSpace(queue)
	if len(queue) == 0 {
		return
	}

	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	connection.queues[queue] = 1
}

func (connection *Connection) UnsubscribeQueue(queue string) {
	queue = strings.TrimSpace(queue)
	if len(queue) == 0 {
		return
	}

	connection.mutex.Lock()
	defer connection.mutex.Unlock()

	delete(connection.queues, queue)
}

func (connection *Connection) Queues() []string {
	var queues []string
	for queue := range connection.queues {
		queues = append(queues, queue)
	}
	return queues
}

func (connection *Connection) IsSubscribedQueue(queue string) bool {
	_, found := connection.queues[queue]
	return found
}

func (connection *Connection) Write(data []byte) (int, error) {
	return connection.client.WriteBytes(data)
}

func (connection *Connection) WriteString(dataString string) (int, error) {
	return connection.client.WriteBytes([]byte(dataString))
}

func (connection *Connection) ResponseError(err string) {
	data, _ := json.Marshal(map[string]interface{}{
		"code":    10000,
		"message": err,
		"data":    nil,
	})
	data = append(data, []byte("\n") ...)
	connection.client.WriteBytes(data)
}

func (connection *Connection) ResponseSuccess(message string) {
	data, _ := json.Marshal(map[string]interface{}{
		"code":    200,
		"message": message,
		"data":    nil,
	})
	data = append(data, []byte("\n") ...)
	connection.client.WriteBytes(data)
}

func (connection *Connection) Close() {
	connection.client.Close()
}

func (connection *Connection) SetWorker(isWorker bool) {
	connection.isWorker = isWorker
}

func (connection *Connection) IsWorker() bool {
	return connection.isWorker
}
