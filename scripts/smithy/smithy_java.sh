#!/usr/bin/env bash
set -e

# JAVA
# Compile library code
cd $FRUGAL_HOME/lib/java && mvn checkstyle:check -q && mvn clean verify -q
mv $(find target -type f -name 'frugal-*.*.*.jar' | grep -v sources | grep -v javadoc) $FRUGAL_HOME

$FRUGAL_HOME/scripts/smithy/codecov.sh $FRUGAL_HOME/lib/java/target/site/jacoco/jacoco.xml java_library
