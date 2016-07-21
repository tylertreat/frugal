#!/usr/bin/env bash
set -ex

cd $GOPATH/src/github.com/Workiva/frugal
frugalDir=$PWD
if [ -f ${frugalDir}/test/integration/log/unexpected_failures.log ]; then cp -r ${frugalDir}/test/integration/log/unexpected_failures.log /testing/artifacts/unexpected_failures.log; fi
tar -czf successful_tests.tar.gz test/integration/log
mv successful_tests.tar.gz /testing/artifacts/
pkill gnatsd
