package master

import (
	"testing"
	"gopkg.in/yaml.v2"
	"os"
)

func TestMaster_Yaml(t *testing.T) {
	type configure struct {
		Master struct {
			Bind string
			Port int
		}
	}
	reader, err := os.Open("/Users/liuxiangchao/Documents/Projects/pp/apps/GoIM/src/main/master/conf/master.conf")
	if err != nil {
		t.Log(err.Error())
		return
	}
	var conf = &configure{}
	yaml.NewDecoder(reader).Decode(conf)

	t.Log(conf)
}

func TestMaster_Start(t *testing.T) {
	NewMaster().StartWithConfig("/Users/liuxiangchao/Documents/Projects/pp/apps/GoIM/src/main/master/conf/master.conf")
}
