#!/bin/sh

# TODO: Poor-man's automated testing, replace this with some proper CI

set -e

TESTDIR=$(dirname "$0")
cd "$TESTDIR/.."
BASEDIR=$(pwd)

echo $TESTDIR
echo $BASEDIR
pwd
printf "Running unit tests:\n"
go test
printf "\nMaking sure all examples work:\n"
for d in $(ls "$BASEDIR/examples"/)
do
    printf "\n * Testing example '%s':\n" "$d"
    go run "$BASEDIR/examples/$d/main.go"
    printf " * Ok\n"
done

printf "\nAll tests Ok\n"
