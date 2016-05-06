#!/bin/sh
PWD=`pwd`
export GOPATH=$PWD:$GOPATH
echo $GOPATH

go build ./src/tvsporedi.si.go

