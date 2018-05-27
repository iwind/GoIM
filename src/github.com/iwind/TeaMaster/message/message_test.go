package message

import (
	"testing"
	"gopkg.in/yaml.v2"
	"log"
	"time"
	"encoding/json"
)

func TestMessage_Yaml(t *testing.T) {
	var data = `
key: user.123
body: 
  name: "Li Bai 李白"
`
	//message := make(map[interface{}]interface{})
	message := struct {
		Key  string
		Body interface{}
	}{}

	t.Log(time.Duration(time.Now().Nanosecond()) / time.Nanosecond)
	err := yaml.Unmarshal([]byte(data), &message)
	if err != nil {
		log.Println(err.Error())
	} else {
		t.Logf("%v\n", message)
	}
	t.Log(time.Duration(time.Now().Nanosecond()) / time.Nanosecond)
}

func TestMessage_JSON(t *testing.T) {
	var data = `{
"key": "user.123",
"body":{
   "name": "Li Bai 李白"
 }
}
`
	//message := make(map[interface{}]interface{})
	message := struct {
		Key  string
		Body interface{}
	}{}

	t.Log(time.Duration(time.Now().Nanosecond()) / time.Nanosecond)
	err := json.Unmarshal([]byte(data), &message)
	if err != nil {
		log.Println(err.Error())
	} else {
		t.Logf("%v\n", message)
	}
	t.Log(time.Duration(time.Now().Nanosecond()) / time.Nanosecond)
}

func TestUnmarshal(t *testing.T) {
	var data = `{
"key": "user.123",
"body":{
   "name": "Li Bai 李白"
 },
 "createdAt": 1527397329.2335
}
`
	message, err := Unmarshal([]byte(data))
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("%#v\n", message)
		t.Logf("%f", message.CreatedAt)
	}
}