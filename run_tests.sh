#!/usr/bin/env bash

which godep > /dev/null || {
    go get github.com/tools/godep
}

# This is so `tee` doesn't absorb a non-zero exit code
set -o pipefail
# Set -e so that we fail if an error is hit.
set -e

thrift -version

# Run tests
cd $GOPATH/src/github.com/Workiva/frugal/test
godep go test -race
