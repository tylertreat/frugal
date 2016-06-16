#!/usr/bin/env bash

set -ex

testing=$PWD
wget -O $GOPATH/bin/thrift https://github.com/stevenosborne-wf/thrift/releases/download/0.9.3-wk-3/thrift-0.9.3-wk-3-linux-amd64
chmod 0755 $GOPATH/bin/thrift		

mkdir -p $GOPATH/src/github.com/Workiva
cp -r $testing $GOPATH/src/github.com/Workiva/frugal

cd $GOPATH/src/github.com/Workiva/frugal
frugal=$PWD
export GOPATH=$GOPATH:$PWD
go get -d ./compiler .
godep go install
go get -d -t ./lib/go .
go build ./lib/go
