package master

import (
	"io/ioutil"
	"log"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaMaster/nets"
	"fmt"
	"sync"
	"github.com/iwind/TeaMaster/node"
	"github.com/iwind/TeaMaster/message"
	"bytes"
	"strconv"
	"net/http"
	"net/url"
	"strings"
	"encoding/json"
)

type Master struct {
	connections      map[int]*Connection    // { ConnectionID: Connection, ... }
	subscriberQueues map[string]map[int]int // { Queue1: [ ConnectionID1:1, ConnectionID2:1, ... ], ... }
	users            map[string]map[int]int // { UserId1: [ ConnectionID1:1, ... ] }
	nodes            []node.Node

	config *Config

	mutex   *sync.Mutex
	idIndex int

	messageHandlers map[string]func(message *message.Message, connection *Connection)
}

type Config struct {
	Bind string
	Port int
	Auth struct {
		On  bool
		API string
	}
}

func NewMaster() *Master {
	master := &Master{
		connections:      map[int]*Connection{},
		subscriberQueues: map[string]map[int]int{},
		users:            map[string]map[int]int{},
		nodes:            []node.Node{},
		mutex:            &sync.Mutex{},
		idIndex:          0,
		messageHandlers:  map[string]func(message *message.Message, connection *Connection){},
	}

	// 内置queue
	// 订阅queue
	master.Handle("$tea.subscribe.queue", func(message *message.Message, connection *Connection) {
		master.mutex.Lock()
		defer master.mutex.Unlock()

		queue, _ := message.StringForKey("queue")
		if len(queue) > 0 {
			if !connection.IsSubscribedQueue(queue) {
				if _, found := master.subscriberQueues[queue]; !found {
					master.subscriberQueues[queue] = map[int]int{}
				}
				master.subscriberQueues[queue][connection.Id()] = 1
				connection.SubscribeQueue(queue)
				connection.ResponseSuccess("ok")
			}
		} else {
			connection.ResponseError("'queue' must not be empty\n")
		}
	})

	// 退出当前连接
	master.Handle("$tea.connection.quit", func(message *message.Message, connection *Connection) {
		connection.Close()
	})

	// 认证
	master.Handle("$tea.connection.auth", func(message *message.Message, connection *Connection) {
		if !master.config.Auth.On {
			connection.ResponseError("Master did not open the authentication")
			return
		}

		token, _ := message.StringForKey("token")
		if len(token) == 0 {
			connection.ResponseError("Need 'body.token' to be a valid string value")
			return
		}

		params := &url.Values{}
		params.Set("TEA_AUTH_TOKEN", token)
		request, err := http.NewRequest(http.MethodPost, master.config.Auth.API, strings.NewReader(params.Encode()))
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
			userIdString := ""
			if ok {
				userIdString = strings.TrimSpace(newUserId)
			} else {
				if userIdInt, ok := userId.(int64); ok {
					userIdString = fmt.Sprintf("%d", userIdInt)
				} else if userIdInt, ok := userId.(int32); ok {
					userIdString = fmt.Sprintf("%d", userIdInt)
				} else if userIdInt, ok := userId.(int16); ok {
					userIdString = fmt.Sprintf("%d", userIdInt)
				} else if userIdInt, ok := userId.(int8); ok {
					userIdString = fmt.Sprintf("%d", userIdInt)
				} else if userIdFloat, ok := userId.(float32); ok {
					userIdString = fmt.Sprintf("%d", int64(userIdFloat))
				} else if userIdFloat, ok := userId.(float64); ok {
					userIdString = fmt.Sprintf("%d", int64(userIdFloat))
				}
			}

			if len(userIdString) == 0 {
				connection.ResponseError("there has a error on authentication server")
				log.Println("Error:" + "'data.userId' should not be empty")
				return
			}

			if len([]byte(userIdString)) > 1024 { // 最大长度不能超过1024
				connection.ResponseError("there has a error on authentication server")
				log.Println("Error:" + "'data.userId' should not be long than 1024 bytes")
				return
			}

			log.Println("register:" + userIdString)
			connection.setUserId(userIdString)

			// 记录到users中
			userConnections, found := master.users[userIdString]
			if !found {
				userConnections = map[int]int{}
				master.users[userIdString] = userConnections
			}
			master.users[userIdString][connection.Id()] = 1

			connection.ResponseSuccess("success")

			return
		}
		connection.ResponseError("there has a error on authentication server")
		log.Println("Error: authenticate server returns a wrong json format:" + string(data))
	})

	return master
}

func (master *Master) StartWithConfig(configFile string) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Failed to start:" + err.Error())
		return
	}

	config := &Config{}
	yaml.Unmarshal(configBytes, config)

	master.config = config

	server := nets.NewServer("tcp", fmt.Sprintf("%s:%d", config.Bind, config.Port))
	server.AcceptClient(func(client *nets.Client) {
		master.mutex.Lock()
		defer master.mutex.Unlock()

		master.idIndex ++
		client.SetId(master.idIndex)

		connection := NewConnection(client)
		master.connections[client.Id()] = connection
	})
	server.CloseClient(func(client *nets.Client) {
		master.mutex.Lock()
		defer master.mutex.Unlock()

		connectionId := client.Id()
		connection, found := master.connections[connectionId]
		if !found {
			return
		}

		log.Println("quit connection:" + strconv.Itoa(connectionId))

		// 从连接列表中删除
		delete(master.connections, connectionId)

		// 从queues中删除
		for _, queue := range connection.Queues() {
			connectionIds, found := master.subscriberQueues[queue]
			if !found {
				continue
			}
			delete(connectionIds, connectionId)
		}

		// 从用户列表中删除
		userId := connection.userId
		if len(userId) > 0 {
			connectionIds, found := master.users[userId]
			if found {
				delete(connectionIds, connectionId)
				if len(connectionIds) == 0 {
					delete(master.users, userId)
				}
			}
		}
	})
	server.ReceiveClient(func(client *nets.Client, data []byte) {
		if len(bytes.TrimSpace(data)) == 0 {
			return
		}

		connection, found := master.connections[client.Id()]
		if !found {
			return
		}

		messageObject, err := message.Unmarshal(data)
		if err != nil {
			connection.ResponseError(err.Error())
			return
		}

		// 判断是否已认证
		if master.config.Auth.On && messageObject.Queue != "$tea.connection.auth" {
			connection.ResponseError("The connection need authenticate")
			return
		}

		if len(messageObject.Queue) > 0 {
			// 是否为内置queue
			handler, found := master.messageHandlers[messageObject.Queue]
			if found {
				handler(messageObject, connection)
			} else {
				messageObject.Pattern = messageObject.Queue

				// 支持具体的queue，如user.1
				connections, found := master.subscriberQueues[messageObject.Queue]
				if found && len(connections) > 0 {
					for connectionId := range connections {
						data, _ := messageObject.Encode()
						master.connections[connectionId].Write(data)
					}
				}

				// @TODO 支持更宽泛的订阅queue，如user.*, user.[1:1000000]

			}
		} else {
			log.Println("Error:Message must has a 'queue'")
			connection.ResponseError("Message must has a 'queue'")
			return
		}
	})
	server.Listen()
}

func (master *Master) Start() {
	configFile := "conf/master.conf"
	master.StartWithConfig(configFile)
}

func (master *Master) Handle(queue string, handler func(message *message.Message, connection *Connection)) {
	master.messageHandlers[queue] = handler
}
