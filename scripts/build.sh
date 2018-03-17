#!/bin/bash

set -e

cd "${0%/*}/.."

out=bin
mkdir -p "$out"

go build -o "$out/gaze" ./src/cmd/gaze

if [[ "$1" == "-r" ]] ; then
    shift
    "$out/gaze" "$@"
fi
