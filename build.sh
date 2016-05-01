#!/bin/sh
PWD=`pwd`
export GOPATH=$PWD:$GOPATH
echo $GOPATH

go build ./src/spored.tv.go

