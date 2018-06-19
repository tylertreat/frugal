#!/usr/bin/env bash
set -e

# Run the generator tests
cd $FRUGAL_HOME
go build -o frugal
go test -race ./test
rm -rf ./test/out
