#!/usr/bin/env bash

export GOPATH=`pwd`/../../

go build -o master/bin/master master.go
cd master
bin/master
