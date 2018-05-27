package mq

import (
	"testing"
	"gopkg.in/yaml.v2"
	"os"
)

func TestMQ_Yaml(t *testing.T) {
	type configure struct {
		Bind string
		Port int
	}
	projectDir, _ := os.LookupEnv("GOPATH")
	reader, err := os.Open(projectDir + "/src/main/mq/conf/mq.conf")
	if err != nil {
		t.Log(err.Error())
		return
	}
	var conf = &configure{}
	yaml.NewDecoder(reader).Decode(conf)

	t.Log(conf)
}

func TestMQ_Start(t *testing.T) {
	projectDir, _ := os.LookupEnv("GOPATH")
	NewMQ().StartWithConfig(projectDir + "/src/main/mq/conf/mq.conf")
}
