# Gaze
Gaze is a "watch" replacement that supports 8bit / 24bit colors.

## Installation

```sh
go install github.com/omakoto/gaze/src/cmd/gaze
```

## Supported options

For now, only the following options are supported.

Short|Long|Description
-----|----|-----------
-n|--interval=FLOAT |Run interval in seconds. 
-p|--precise|Attempt run command in precise intervals.
-r|--repeat=N|Repeat command N times and finish.
-t|--no-title|Turn off header.
-x|--exec|Pass command to exec instead of "sh -c".
-c|--color|Ignored. ANSI colors are always preserved.

## Unsupported options

The following options from GNU watch are not supported yet.

Short|Long|Description
-----|----|-----------
-b|--beep|Beep if command has a non-zero exit.
-d|--differences[=permanent]|Highlight changes between updates.
-e|--errexit|Exit if command has a non-zero exit.
-g|--chgexit|Exit when output from command changes.

## TODOs
 - It somehow breaks alternate screen buffer (e.g. fzf)
