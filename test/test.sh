#!/bin/sh

# TODO: Poor man's automated testing, replace this with some proper CI

set -e

TESTDIR=$(dirname "$0")
cd "$TESTDIR/.."
BASEDIR=$(pwd)

printf "Running unit tests:\n"
go test -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
printf "See coverage.html for more details\n"
printf "\nMaking sure all examples work:\n"
for d in $(ls "$BASEDIR/examples"/)
do
    printf "\n * Testing example '%s':\n" "$d"
    go run "$BASEDIR/examples/$d/main.go"
    printf " * Ok\n"
done

printf "\nAll tests Ok\n"
