package mq

import "testing"

func TestConnection_Keys(t *testing.T) {
	var connection = NewConnection(nil)
	connection.SubscribeQueue("key.hello")
	connection.SubscribeQueue("key.world")
	connection.SubscribeQueue("key.user.1")
	t.Log(connection.Queues())

	connection.UnsubscribeQueue("key.world")
	t.Log(connection.Queues())

	connection.UnsubscribeQueue("")
	t.Log(connection.Queues())
}

