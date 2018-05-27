#!/usr/bin/env bash

export GOPATH=`pwd`/../../

go build -o worker/bin/worker worker.go
cd worker
bin/worker