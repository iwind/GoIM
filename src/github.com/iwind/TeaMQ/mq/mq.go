package mq

import (
	"io/ioutil"
	"log"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaMQ/nets"
	"fmt"
	"sync"
	"github.com/iwind/TeaMQ/message"
	"bytes"
	"strconv"
	"net/http"
	"net/url"
	"strings"
	"encoding/json"
	"github.com/iwind/TeaMQ/worker"
)

type MQ struct {
	connections      map[int]*Connection    // { ConnectionID: Connection, ... }
	subscriberQueues map[string]map[int]int // { Queue1: [ ConnectionID1:1, ConnectionID2:1, ... ], ... }
	users            map[int64]map[int]int  // { UserId1: [ ConnectionID1:1, ... ] }
	workers          map[int]*worker.Worker // { ConnectionID: Work1, ... }

	config *Config

	mutex   *sync.Mutex
	idIndex int

	messageHandlers map[string]func(message *message.Message, connection *Connection)
}

type Config struct {
	Bind string
	Port int
	Keys []string
	Auth struct {
		On  bool
		API string
	}
}

func NewMQ() *MQ {
	mq := &MQ{
		connections:      map[int]*Connection{},
		subscriberQueues: map[string]map[int]int{},
		users:            map[int64]map[int]int{},
		workers:          map[int]*worker.Worker{},
		mutex:            &sync.Mutex{},
		idIndex:          0,
		messageHandlers:  map[string]func(message *message.Message, connection *Connection){},
	}

	// 处理内置queue
	mq.Handle("$tea.subscribe.queue", func(message *message.Message, connection *Connection) {
		mq.mutex.Lock()
		defer mq.mutex.Unlock()

		queue, _ := message.StringForKey("queue")
		if len(queue) > 0 {
			if !connection.IsSubscribedQueue(queue) {
				if _, found := mq.subscriberQueues[queue]; !found {
					mq.subscriberQueues[queue] = map[int]int{}
				}
				mq.subscriberQueues[queue][connection.Id()] = 1
				connection.SubscribeQueue(queue)
				connection.ResponseSuccess("ok")
			}
		} else {
			connection.ResponseError("'queue' must not be empty\n")
		}
	})

	// 退出当前连接
	mq.Handle("$tea.connection.quit", func(message *message.Message, connection *Connection) {
		connection.Close()
	})

	// 认证
	mq.Handle("$tea.connection.auth", func(message *message.Message, connection *Connection) {
		if !mq.config.Auth.On {
			connection.ResponseError("MQ did not open the authentication")
			return
		}

		token, _ := message.StringForKey("token")
		if len(token) == 0 {
			connection.ResponseError("Need 'body.token' to be a valid string value")
			return
		}

		params := &url.Values{}
		params.Set("TEA_AUTH_TOKEN", token)
		request, err := http.NewRequest(http.MethodPost, mq.config.Auth.API, strings.NewReader(params.Encode()))
		if err != nil {
			connection.ResponseError(err.Error())
		}
		request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			connection.ResponseError("there has a error on authentication server")
			log.Println("Error:" + err.Error())
			return
		}

		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			connection.ResponseError("there has a error on authentication server")
			log.Println("Error:" + err.Error())
			return
		}

		responseJSON := &struct {
			Code    int
			Message string
			Data    map[string]interface{}
		}{}
		err = json.Unmarshal(data, responseJSON)
		if err != nil {
			connection.ResponseError("there has a error on authentication server")
			log.Println("Error:" + err.Error())
			return
		}

		if responseJSON.Code == 200 {
			if responseJSON.Data == nil {
				connection.ResponseError("there has a error on authentication server")
				log.Println("Error:" + "'data' should be in valid format")
				return
			}

			userId, found := responseJSON.Data["userId"]
			if !found {
				connection.ResponseError("there has a error on authentication server")
				log.Println("Error:" + "'data.userId' should be in valid format")
				return
			}

			newUserId, ok := userId.(string)
			realUserId := int64(0)
			if ok {
				userIdString := strings.TrimSpace(newUserId)
				userIdInt, err := strconv.Atoi(userIdString)
				if err != nil {
					connection.ResponseError("there has a error on authentication server")
					log.Println("Error:" + "'data.userId' should not be a integer number")
					return
				}
				realUserId = int64(userIdInt)
			} else {
				if userIdInt, ok := userId.(int64); ok {
					realUserId = userIdInt
				} else if userIdInt, ok := userId.(int32); ok {
					realUserId = int64(userIdInt)
				} else if userIdInt, ok := userId.(int16); ok {
					realUserId = int64(userIdInt)
				} else if userIdInt, ok := userId.(int8); ok {
					realUserId = int64(userIdInt)
				} else if userIdFloat, ok := userId.(float32); ok {
					realUserId = int64(userIdFloat)
				} else if userIdFloat, ok := userId.(float64); ok {
					realUserId = int64(userIdFloat)
				}
			}

			log.Printf("register:%d\n", realUserId)
			connection.setUserId(realUserId)

			// 记录到users中
			userConnections, found := mq.users[realUserId]
			if !found {
				userConnections = map[int]int{}
				mq.users[realUserId] = userConnections
			}
			mq.users[realUserId][connection.Id()] = 1

			connection.ResponseSuccess("ok")

			return
		}
		connection.ResponseError("there has a error on authentication server")
		log.Println("Error: authenticate server returns a wrong json format:" + string(data))
	})

	// 注册Worker
	mq.Handle("$tea.worker.register", func(message *message.Message, connection *Connection) {
		log.Println("Register new worker")

		// 检查Key
		key := message.StringForKeyDefault("key", "")
		if len(key) == 0 {
			connection.ResponseError("Register failed, key must be specified")
			return
		}
		found := false
		for _, savedKey := range mq.config.Keys {
			if savedKey == key {
				found = true
				break
			}
		}
		if !found {
			connection.ResponseError("Register failed, key '" + key + "' is invalid")
			return
		}

		mq.mutex.Lock()
		defer mq.mutex.Unlock()

		workerObject := worker.NewWorker()
		workerObject.Id = message.StringForKeyDefault("id", "")
		workerObject.Name = message.StringForKeyDefault("name", "")
		workerObject.Description = message.StringForKeyDefault("description", "")
		workerObject.Key = message.StringForKeyDefault("key", "")

		userMap, found := message.MapForKey("user")
		if found {
			minValue, found := userMap["min"]
			if found {
				minValueFloat, ok := minValue.(float64)
				if ok {
					workerObject.User.Min = int64(minValueFloat)
				}
			}

			maxValue, found := userMap["max"]
			if found {
				maxValueFloat, ok := maxValue.(float64)
				if ok {
					workerObject.User.Max = int64(maxValueFloat)
				}
			}
		}

		mq.workers[connection.Id()] = workerObject

		connection.ResponseSuccess("ok")
	})

	return mq
}

func (mq *MQ) StartWithConfig(configFile string) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Failed to start:" + err.Error())
		return
	}

	config := &Config{}
	yaml.Unmarshal(configBytes, config)

	mq.config = config

	server := nets.NewServer("tcp", fmt.Sprintf("%s:%d", config.Bind, config.Port))
	server.AcceptClient(func(client *nets.Client) {
		mq.mutex.Lock()
		defer mq.mutex.Unlock()

		mq.idIndex ++
		client.SetId(mq.idIndex)

		connection := NewConnection(client)
		mq.connections[client.Id()] = connection

		log.Printf("Accept new connection %d\n", client.Id())

		//@TODO 30秒内没有认证自动关闭
		if mq.config.Auth.On {

		}
	})
	server.CloseClient(func(client *nets.Client) {
		mq.mutex.Lock()
		defer mq.mutex.Unlock()

		connectionId := client.Id()
		connection, found := mq.connections[connectionId]
		if !found {
			return
		}

		log.Println("quit connection " + strconv.Itoa(connectionId))

		// 从连接列表中删除
		delete(mq.connections, connectionId)

		// 从queues中删除
		for _, queue := range connection.Queues() {
			connectionIds, found := mq.subscriberQueues[queue]
			if !found {
				continue
			}
			delete(connectionIds, connectionId)
		}

		// 从用户列表中删除
		userId := connection.userId
		if userId > 0 {
			connectionIds, found := mq.users[userId]
			if found {
				delete(connectionIds, connectionId)
				if len(connectionIds) == 0 {
					log.Printf("Remove user %d\n", userId)
					delete(mq.users, userId)
				}
			}
		}

		// 从workers中删除
		if _, ok := mq.workers[connectionId]; ok {
			log.Println("Remove worker " + strconv.Itoa(connectionId))
			delete(mq.workers, connectionId)
		}
	})
	server.ReceiveClient(func(client *nets.Client, data []byte) {
		if len(bytes.TrimSpace(data)) == 0 {
			return
		}

		connection, found := mq.connections[client.Id()]
		if !found {
			return
		}

		messageObject, err := message.Unmarshal(data)
		if err != nil {
			connection.ResponseError(err.Error())
			return
		}

		// 判断是否已认证
		if mq.config.Auth.On && messageObject.Queue != "$tea.connection.auth" && messageObject.Queue != "$tea.worker.register" {
			connection.ResponseError("The connection need authenticate")
			return
		}

		if len(messageObject.Queue) > 0 {
			// 是否为内置queue
			handler, found := mq.messageHandlers[messageObject.Queue]
			if found {
				handler(messageObject, connection)
			} else {
				// 非内置queue
				// 如果是来自worker，则直接发送到用户端
				if connection.isWorker {
					messageObject.Pattern = messageObject.Queue

					// 支持具体的queue，如user.1
					connections, found := mq.subscriberQueues[messageObject.Queue]
					if found && len(connections) > 0 {
						for connectionId := range connections {
							data, err := messageObject.Encode()
							if err != nil {
								log.Println("Error:" + err.Error())
							} else {
								mq.connections[connectionId].Write(data)
							}
						}
					}

					// @TODO 支持更宽泛的订阅queue，如user.*, user.[1:1000000]
				} else {
					log.Println("receive " + string(data))

					// 如果来自用户端，则转发到worker
					selectedConnectionId := 0
					if messageObject.FromUserId() > 0 && len(mq.workers) > 0 {
						for workerConnectionId, workerObject := range mq.workers {
							if workerObject.User.Min <= messageObject.FromUserId() && workerObject.User.Max >= messageObject.FromUserId() {
								selectedConnectionId = workerConnectionId
							}
						}
					}
					if selectedConnectionId == 0 && len(mq.workers) > 0 {
						// 随机选一个
						for id := range mq.workers {
							selectedConnectionId = id
							break
						}
					}

					if selectedConnectionId == 0 {
						log.Println("Error:There is not worker for the message")
					} else {
						connection, found := mq.connections[selectedConnectionId]
						if found {
							data, err := messageObject.Encode()
							if err != nil {
								log.Println("Error:" + err.Error())
							} else {
								connection.Write(data)
							}
						}
					}
				}

			}
		} else {
			log.Println("Error:Message must has a 'queue'")
			connection.ResponseError("Message must has a 'queue'")
			return
		}
	})
	server.Listen()
}

func (mq *MQ) Start() {
	configFile := "conf/mq.conf"
	mq.StartWithConfig(configFile)
}

func (mq *MQ) Handle(queue string, handler func(message *message.Message, connection *Connection)) {
	mq.messageHandlers[queue] = handler
}
