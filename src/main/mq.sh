#!/usr/bin/env bash

export GOPATH=`pwd`/../../

go build -o mq/bin/mq mq.go
cd mq
bin/mq
