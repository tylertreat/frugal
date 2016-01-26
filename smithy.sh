#!/usr/bin/env bash

# Get maven
wget http://apache.claz.org/maven/maven-3/3.3.9/binaries/apache-maven-3.3.9-bin.tar.gz
tar xzvf apache-maven-3.3.9-bin.tar.gz
export PATH=$PWD/apache-maven-3.3.9/bin:$PATH

# Compile the java library code
ROOT=$PWD
cd $ROOT/lib/java && mvn verify
mv target/frugal-*.jar $ROOT
cd $ROOT

# Run the generator tests
go get -d ./compiler .
go build -o frugal
# go test ./test
# rm -rf ./test/out
