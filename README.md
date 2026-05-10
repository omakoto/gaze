# Gaze

Gaze is a `watch` replacement that preserves 8-bit and 24-bit ANSI colors in command output. It repeatedly runs a command and displays the result, with interactive controls for pausing, refreshing, and adjusting the interval on the fly.

## Installation

```bash
go install github.com/omakoto/gaze/src/cmd/gaze@latest
```

## Usage

```
gaze [OPTIONS] COMMAND [ARGS...]
```

**Examples:**

```bash
# Run every 2 seconds (default)
gaze ls -la

# Run every 0.5 seconds
gaze -n 0.5 cat /proc/loadavg

# Run exactly 10 times, then exit
gaze -r 10 date

# Pass a shell pipeline
gaze 'ps aux | grep python'

# Skip the header line
gaze -t df -h
```

## Options

Short | Long | Description
------|------|------------
`-n` | `--interval=SECONDS` | Repeat interval in seconds (default: 2.0, minimum: 0.1)
`-p` | `--precise` | Use precise intervals: measure from start of run, not end
`-r` | `--repeat=N` | Run command N times, then exit
`-t` | `--no-title` | Hide the header line
`-x` | `--exec` | Run command directly via exec(2) instead of `sh -c`
`-c` | `--color` | Ignored — ANSI colors are always preserved
    | `--width=N` | Override terminal width (default: auto-detect)
    | `--height=N` | Override terminal height (default: auto-detect)

### Precise vs. normal interval timing

By default, the interval is measured from when a command finishes to when the next one starts, so slow commands space out runs further. With `--precise`, gaze targets consistent wall-clock start times regardless of how long the command takes.

## Interactive controls

While gaze is running, you can use these keys:

Key | Action
----|-------
`Enter` | Force an immediate refresh
`Space` | Toggle pause / resume
`+` | Increase interval by 0.5s
`-` | Decrease interval by 0.5s
`q` | Quit

When paused, `Enter` triggers a single manual refresh without resuming auto-refresh.

## Comparison with GNU watch

Gaze supports ANSI colors natively (no flags needed). The following GNU watch options are **not yet implemented**:

Short | Long | Description
------|------|------------
`-b` | `--beep` | Beep on non-zero exit
`-d` | `--differences` | Highlight changes between updates
`-e` | `--errexit` | Exit on non-zero exit
`-g` | `--chgexit` | Exit when output changes
