package TeaDemo

import (
	"github.com/iwind/TeaWorker/worker"
	"github.com/iwind/TeaDemo/messages"
)

func Start() {
	worker.
		NewWorker().
		Handle("GET_USER_PROFILE", messages.GetUserProfile).
		Handle("UPDATE_USER_NAME", messages.UpdateUserName).
		Start()
}
