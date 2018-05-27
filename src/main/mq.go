package main

import "github.com/iwind/TeaMQ/mq"

func main() {
	mq.NewMQ().Start()
}