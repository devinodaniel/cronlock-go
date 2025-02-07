# cronlock-go

Distributed cron lock written in go with a UI for visualizing current crons and their status.

But why?

Features:

- run multiple cron servers for high availability
- easily perform maintenance on cron hosts without missing crons
- expose metrics like how many crons are running, their pass/fail statuses, and duration  

## Install Redis

You must have a running Redis server for cronlock to work.

### MacOS:

`brew install redis`

Start the redis server:

`brew services start redis`

Stop the redis server:

`brew services stop redis`

## Build

Generate the `cronlock` and `cronlockweb` binaries.

`make`

# Run a command

Without cron:

```
./cronlock sleep 60
```

With cron:

```bash
* * * * * cronlock sleep 60
```

## Start the web UI

`make web`

### Templates

HTML templates (`cmd/cronlockweb/templates`) should be placed in the working directory of `cronlockweb`.

## Configuration

Any of these options can be passed to `cronlock` per command or set as global environment variables.

`CRONLOCK_REDIS_HOST` - (default: localhost) hostname or IP address of the Redis server

`CRONLOCK_REDIS_PORT` - (default: 6379) port Redis is running on

`CRONLOCK_TIMEOUT`- (default: 3600) expire the cronlock after N seconds if it somehow fails to report success or failure. 0 creates permanent lock for crons with the same hash. useful to ensure a cron never gets ran more than once

`CRONLOCK_GRACE_PERIOD` - (default: 5 sec) remove lock N seconds after cron ends. useful for quick crons. must be greater than 0. When `CRONLOCK_TIMEOUT` is set to zero, this has no use

For debugging and more verbose feedback:

`CRONLOCK_DEBUG` - (default: false) show debug log lines

`CRONLOCK_PRINT_STDOUT` - (default: false) print stdout of command when executing cron

`CRONLOCK_PRINT_ARGS` - (default: false) print the script arguments when executing cron

## Tests

Run all the tests:

`make tests`

## Cleanup

Remove the binaries

`make clean`

## Remove all crons from Redis

To start from a clean slate, destroy all cron locks in Redis.

`make flush`
