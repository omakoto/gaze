#!/bin/bash

set -e
cd "$(dirname "$0")"

. ./common.bash

../bin/gaze --repeat 1 --width 40 --height 5 yes
