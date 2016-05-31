#!/usr/bin/env bash

set -ex

# Grab thrift binary
mkdir -p $PWD/bin
wget -O $PWD/bin/thrift https://github.com/stevenosborne-wf/thrift/releases/download/0.9.3-wk-2/thrift-0.9.3-wk-2-linux-amd64
chmod 0755 $PWD/bin/thrift
export PATH=$PATH:$PWD/bin
testing=$PWD

mkdir -p $GOPATH/src/github.com/Workiva
cp -r $testing $GOPATH/src/github.com/Workiva/frugal

cd $GOPATH/src/github.com/Workiva/frugal
frugal=$PWD
export GOPATH=$GOPATH:$PWD
go get -d ./compiler .
go install
go get -d -t ./lib/go .
go build ./lib/go