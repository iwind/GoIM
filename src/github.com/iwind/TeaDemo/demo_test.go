package TeaDemo

import (
	"testing"
	"os"
	"github.com/iwind/TeaDemo/messages"
	"github.com/iwind/TeaWorker/worker"
)

func TestWorker_Start(t *testing.T) {
	projectDir, _ := os.LookupEnv("GOPATH")
	worker.
		NewWorker().
		Handle("GET_USER_PROFILE", messages.GetUserProfile).
		Handle("UPDATE_USER_NAME", messages.UpdateUserName).
		StartWithConfig(projectDir + "/src/main/worker/conf/worker.conf")
}

