#!/bin/bash

set -e
cd "$(dirname "$0")"

. ./common.bash

../bin/gaze --repeat 1 --width 41 --height 5 'while true; do echo -n "あ" ; done'