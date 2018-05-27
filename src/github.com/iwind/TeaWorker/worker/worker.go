package worker

import (
	"github.com/iwind/TeaWorker/message"
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
	"github.com/iwind/TeaWorker/nets"
	"fmt"
	"time"
)

type Worker struct {
	handlers map[string]func(message *message.Message, worker *Worker)
}

type Config struct {
	MQ struct {
		Host string
		Port int
	}
	Id          string
	Name        string
	Description string
	Key         string
	User struct {
		Min int64
		Max int64
	}
}

func NewWorker() *Worker {
	worker := &Worker{
		handlers: map[string]func(message *message.Message, worker *Worker){},
	}

	return worker
}

func (worker *Worker) Handle(messageType string, handler func(message *message.Message, worker *Worker)) (*Worker) {
	worker.handlers[messageType] = handler
	return worker
}

func (worker *Worker) Subscribe(queue string) *Worker {
	return worker
}

func (worker *Worker) Start() {
	worker.StartWithConfig("conf/worker.conf")
}

func (worker *Worker) StartWithConfig(configFile string) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Println("Error:" + err.Error())
		return
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		log.Println("Error:" + err.Error())
		return
	}

	for {
		// 连接MQ
		client := &nets.Client{}
		err = client.Connect("tcp", fmt.Sprintf("%s:%d", config.MQ.Host, config.MQ.Port))
		if err != nil {
			log.Println("Error:" + err.Error())

			time.Sleep(time.Second * 5)

			continue
		}

		// 注册Work，包括用户range
		messageObject := message.NewMessage()
		messageObject.Queue = "$tea.worker.register"
		messageObject.Set("id", config.Id)
		messageObject.Set("name", config.Name)
		messageObject.Set("description", config.Description)
		messageObject.Set("key", config.Key)
		messageObject.Set("user", map[string]int64{
			"min": config.User.Min,
			"max": config.User.Max,
		})
		data, err := messageObject.Encode()
		if err != nil {
			log.Println("Error:" + err.Error())
			return
		}
		log.Println("prepare to send data:" + string(data))
		client.WriteBytes(data)

		// 接收数据
		client.Receive(func(data []byte) {
			log.Println(string(data))
		})
	}
}
