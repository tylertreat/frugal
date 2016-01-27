#!/usr/bin/env bash

set -eo pipefail

# Set the outfile
outfile=$PWD/gotest.out

echo "Running integration tests"
godep go test -race -v github.com/Workiva/frugal/test/integration | tee $outfile

# Convert the out file to xml
go2xunit -input $outfile -output /testing/reports/integration_tests.xml
