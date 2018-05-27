package message

import (
	"encoding/json"
	"errors"
	"time"
	"fmt"
	"math/rand"
)

type Message struct {
	id         string
	isSent     bool
	isReceived bool

	Queue      string
	Pattern    string
	Body       map[string]interface{}
	CreatedAt  float64
	sentAt     float64
	receivedAt float64
}

func Unmarshal(data []byte) (*Message, error) {
	messageMap := map[string]interface{}{}

	err := json.Unmarshal(data, &messageMap)
	if err != nil {
		return nil, err
	}

	message := &Message{
		isSent:     false,
		isReceived: false,
	}

	// id
	id, found := messageMap["id"]
	if found {
		if idString, ok := id.(string); ok {
			message.id = idString
		}
	}

	// Queue
	queue, found := messageMap["queue"]
	if !found {
		return nil, errors.New("message should contains 'queue' field")
	}

	queueString, ok := queue.(string)
	if !ok {
		return nil, errors.New("message queue should be a string")
	}
	message.Queue = queueString

	// Body
	body, found := messageMap["body"]
	if found {
		bodyMap, ok := body.(map[string]interface{})
		if ok {
			message.Body = bodyMap
		}
	}

	// CreatedAt
	createdAt, found := messageMap["createdAt"]
	if found {
		createdAtFloat, ok := createdAt.(float64)
		if ok {
			message.CreatedAt = createdAtFloat
		}
	}

	return message, nil
}

func (message *Message) Encode() ([]byte, error) {
	messageJSON := map[string]interface{}{
		"id":        message.id,
		"queue":     message.Queue,
		"createdAt": message.CreatedAt,
		"body":      message.Body,
		"meta": map[string]interface{}{
			"uniqueId": fmt.Sprintf("%d%d", time.Now().Nanosecond(), rand.NewSource(time.Now().UnixNano()).Int63())[:32],
			"pattern":  message.Pattern,
			"sentAt":   float64(time.Now().UnixNano()) / 1000000000,
		},
	}
	data, err := json.Marshal(messageJSON)
	data = append(data, []byte("\n")...)
	return data, err
}

func (message *Message) ValueForKey(key string) interface{} {
	if message.Body != nil {
		value, found := message.Body[key]
		if !found {
			return nil
		}
		return value
	}
	return nil
}

func (message *Message) StringForKey(key string) (string, bool) {
	value := message.ValueForKey(key)
	if value == nil {
		return "", false
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", false
	}

	return stringValue, true
}
