package message

import "encoding/json"

type Message struct {
	Queue string
	Body  map[string]interface{}
}

func NewMessage() *Message {
	return &Message{
		Queue: "",
		Body:  map[string]interface{}{},
	}
}

func (message *Message) Set(key string, value interface{}) {
	message.Body[key] = value
}

func (message *Message) Encode() ([]byte, error) {
	data, err := json.Marshal(map[string]interface{}{
		"queue": message.Queue,
		"body":  message.Body,
	})
	if err == nil {
		data = append(data, []byte("\n")...)
	}
	return data, err
}
