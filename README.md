# cronlock-go

Distributed cron lock written in go with a UI for visualizing current crons and their status.

But why?

This allow you to have multiple cron servers running the same crons for high availability. Cronlock will prevent a cron from being executed on multiple hosts. The first and fastest cron host gets to execute the cron. You can easily upgrade or perform maintenance on one or more crons and still have reliable crons.

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

HTML templates should be in the working directory of the `./cronlockweb` binary.

## Tests

Run all the tests:

`make tests`

## Cleanup

Remove the binaries

`make clean`

## Remove all crons from Redis

To start from a clean slate, destroy all cron locks in Redis.

`make flush`
