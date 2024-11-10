#!/bin/sh

set -e

TESTDIR=$(realpath "$(dirname "$0")")
BASEDIR=$(dirname "$TESTDIR")

printf "Making sure all examples work:\n"
for d in $(ls "$BASEDIR/examples"/)
do
    printf " * Testing example '%s':\n" "$d"
    go run "$BASEDIR/examples/$d/main.go"
    printf " * Ok\n\n"
done
