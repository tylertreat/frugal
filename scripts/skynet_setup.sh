#!/usr/bin/env bash

set -ex

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
