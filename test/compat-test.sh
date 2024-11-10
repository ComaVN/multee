#!/bin/sh

set -e

TESTDIR=$(realpath "$(dirname "$0")")
BASEDIR=$(dirname "$TESTDIR")

printf "Running Go version compatibility tests:\n"
for v in 1.21 1.22 1.23
do
    docker build --build-arg GO_VERSION="$v" .
done
