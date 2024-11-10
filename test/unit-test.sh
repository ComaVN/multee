#!/bin/sh

set -e

TESTDIR=$(realpath "$(dirname "$0")")
BASEDIR=$(dirname "$TESTDIR")

printf "Running unit tests:\n"
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
printf "See coverage.html for more details\n\n"
