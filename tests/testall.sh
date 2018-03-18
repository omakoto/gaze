#!/bin/bash

here=$(dirname "$0")

filter="$*"
: ${filter:='*'}

if [[ "$WITH_DEBUGGER" == 1 ]] ; then
    debug=--debug
fi

export debug

export bin=../bin/gaze

die() {
  echo "$@"
  exit 2
}

cd $here || die "$0: can't chdir to $here."

num_pass=0
num_fail=0

../scripts/build.sh || exit 1

for test in *.t ; do
  name=$(basename $test .t)

  echo -n "Test: $name "
  expect="${name}.expected"
  actual="${name}.actual"
  diff="${name}.diff"
  bash $test >"$actual"
  diff --color=never -u "$expect" "$actual" >"$diff"
  rc=$?
  if (( $rc == 0 )) ; then
    echo $'\e[32;1mpass\e[0m'
    num_pass=$(( $num_pass + 1 ))
  else
    echo $'\e[31;1mFAIL\e[0m'
    num_fail=$(( $num_fail + 1 ))
  fi
done

if (( $num_pass > 0 && $num_fail == 0 )) ; then
    exit 0
else
    exit 1
fi
