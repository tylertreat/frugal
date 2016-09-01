#!/usr/bin/env bash

set -exo pipefail

# Move godeps to gopath for both library and binary
cp -r /testing/Godeps/_workspace/* $GOPATH/
cp -r /testing/lib/go/Godeps/_workspace/* $GOPATH/

# Move frugal itself into gopath
mkdir -p $GOPATH/src/github.com/Workiva
cp -r /testing $GOPATH/src/github.com/Workiva/frugal
cd $GOPATH/src/github.com/Workiva/frugal

# Install frugal
godep go install

# Start gnatsd
gnatsd &

