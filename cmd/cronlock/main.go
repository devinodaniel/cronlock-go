package main

// CronLock is a simple distributed lock for cron jobs.
// It is designed to be used with a distributed key-value store, such as Redis.
// It is based off a bash script by the same name at https://github.com/kvz/cronlock

import (
	"os"

	"github.com/devinodaniel/cronlock-go/common/config"
	"github.com/devinodaniel/cronlock-go/common/cron"
	"github.com/devinodaniel/cronlock-go/common/log"
	"github.com/devinodaniel/cronlock-go/common/redis"
)

var (
	err error
)

func main() {
	// get the arguments passed to the script
	args := os.Args[1:]

	// create a new cron object
	cron := cron.New(args)

	// open redis connection
	cron.RedisClient, err = redis.Connect(config.CRONLOCK_REDIS_HOST, config.CRONLOCK_REDIS_PORT, config.CRONLOCK_REDIS_DATABASE)
	if err != nil {
		log.Fatal("Failed to connect to Redis: %v", err)
	}

	// explicitly close the redis connection when the script finishes
	defer cron.RedisClient.Close()

	// run the cron job
	if err := cron.Run(); err != nil {
		log.Error("%v", err)
	}
}
