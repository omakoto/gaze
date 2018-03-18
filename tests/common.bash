time_file=/tmp/$$.time
echo 1500000000 >$time_file

export GAZE_TIME_INJECTION_FILE=$time_file

export TZ='America/Los_Angeles'
