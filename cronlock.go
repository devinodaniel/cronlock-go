package main

// CronLock is a simple distributed lock for cron jobs.
// It is designed to be used with a distributed key-value store, such as Redis.
// It is based off a bash script by the same name at https://github.com/kvz/cronlock

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"

	redis "github.com/redis/go-redis/v9"
)

var (
	err           error
	logDateFormat = "2006-01-02 15:04:05"
)

// configuration defaults (can be overridden by environment variables)
var (
	CRONLOCK_HOST           = getEnvStr("CRONLOCK_HOST", "localhost")
	CRONLOCK_PORT           = getEnvInt("CRONLOCK_PORT", 6379)
	CRONLOCK_RETRY_ATTEMPTS = getEnvInt("CRONLOCK_RETRY_ATTEMPTS", 5)
	CRONLOCK_DEBUG          = getEnvStr("CRONLOCK_DEBUG", "false")
	CRONLOCK_KEEP_HISTORY   = getEnvStr("CRONLOCK_KEEP_HISTORY", "false")
	CRONLOCK_EXPIRY_TIME    = getEnvInt("CRONLOCK_EXPIRY_TIME", 86400)
)

const (
	CRON_STATUS_RUNNING  = "RUNNING"
	CRON_STATUS_SUCCESS  = "SUCCESS"
	CRON_STATUS_FAILED   = "FAILED"
	CRON_STATUS_COMPLETE = "COMPLETE"
)

type Cron struct {
	RedisClient *redis.Client
	Ctx         context.Context
	EpochStart  int64    `json:"epochStart"`
	EpochEnd    int64    `json:"epochEnd"`
	Duration    int64    `json:"duration"`
	Md5Hash     string   `json:"md5Hash"`
	Args        []string `json:"args"`
	Status      string   `json:"status"`
}

func main() {
	// get the arguments passed to the script
	args := os.Args[1:]

	// create a new cron object
	cron := Cron{
		Args:    args,
		Md5Hash: hash(args),
		Ctx:     context.Background(),
	}

	// open redis connection
	cron.RedisClient, err = connectWithRetries()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}

	// explicitly close the redis connection when the script finishes
	defer cron.RedisClient.Close()

	// run the cron job
	if err := cron.Run(); err != nil {
		log.Fatalf("%v\n", err)
	}
}

func (cron *Cron) Run() error {
	// run the command and set the metadata
	if err := cron.start(); err != nil {
		cron.Status = CRON_STATUS_FAILED
		return err
	}

	if cron.Status == CRON_STATUS_SUCCESS {
		// update the metadata after the command has run
		if err := cron.complete(); err != nil {
			return err
		}
	}

	return nil
}

// start() runs the command and sets the metadata if the command is not already running
func (cron *Cron) start() error {
	// set the start time
	cron.EpochStart = time.Now().Unix()
	// set the status to running
	cron.Status = CRON_STATUS_RUNNING
	// expire the key after 24 hours in case the command fails to finish
	expiryTime := time.Duration(CRONLOCK_EXPIRY_TIME) * time.Second

	// check if the command is already running by setting the key
	// SetNX returns true if the key was set, false if it already exists
	okToRun, err := cron.RedisClient.SetNX(cron.Ctx, cron.Md5Hash, cron, expiryTime).Result()
	if err != nil {
		return err
	}

	// if the key was already set, the command is already running
	if !okToRun {
		logInfo("%s already running. Skipping.", cron.Md5Hash)
		return nil
	}

	// if the key was set, the command is ok to run
	if okToRun {
		logDebug("%s added to Redis", cron.Md5Hash)
		logInfo("%s starting", cron.Md5Hash)

		// run the command
		if err := raw_cmd(cron.Args); err != nil {
			return err
		}

		cron.Status = CRON_STATUS_SUCCESS
	}

	return nil
}

// finished() updates the metadata after the command has run
func (cron *Cron) complete() error {
	cron.EpochEnd = time.Now().Unix()
	cron.Duration = cron.EpochEnd - cron.EpochStart

	var expiryTime time.Duration

	switch CRONLOCK_KEEP_HISTORY {
	case "true":
		// set expiration time to keep forever (0)
		expiryTime = time.Duration(0) * time.Second
	case "false":
		// set expiration time to 10 seconds
		expiryTime = time.Duration(20) * time.Second
	}

	cron.Status = CRON_STATUS_COMPLETE

	// update the metadata
	updateResult, err := cron.RedisClient.Set(cron.Ctx, cron.Md5Hash, cron, expiryTime).Result()
	if err != nil {
		return err
	}
	if updateResult == "OK" {
		logDebug("%s metadata updated in Redis", cron.Md5Hash)
	}

	logInfo("%s finished after %d seconds", cron.Md5Hash, cron.Duration)

	return nil
}

// MarshalBinary automatically marshals the cron object into a JSON object to be stored in Redis
func (cron Cron) MarshalBinary() ([]byte, error) {
	// Custom logic to marshal your type into JSON
	bytes, err := json.Marshal(cron)
	return bytes, err
}

// getEnvString retrieves the string value of the environment variable named by the key.
func getEnvStr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	} else {
		return defaultValue
	}

}

// getEnv retrieves the integer value of the environment variable named by the key.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// make sure that the value is an integer
		value, _ := strconv.Atoi(value)
		return value
	} else {
		return defaultValue
	}
}

// hash creates an md5 hash of all the arugments passed to it
func hash(args ...interface{}) string {
	// return the md5 hash of the arguments
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprint(args...))))
}

// connect() makes a connection to redis and returns a client
func connect() *redis.Client {
	// connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", CRONLOCK_HOST, CRONLOCK_PORT),
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

	return client
}

// connectWithRetries() makes a connection to redis and retries if it fails
func connectWithRetries() (*redis.Client, error) {
	for connectAttempt := 0; connectAttempt < CRONLOCK_RETRY_ATTEMPTS; connectAttempt++ {
		client := connect()
		if client != nil {
			return client, nil
		}
		// sleep 1 second before retrying connection if it failed
		time.Sleep(1 * time.Second)
		logDebug("Connection failed, retrying... attempt %d\n", connectAttempt+1)
	}
	return nil, fmt.Errorf("Failed to connect to Redis after %d attempts", CRONLOCK_RETRY_ATTEMPTS)
}

// raw_cmd() executes the cron commands or scripts and returns an error if it fails
func raw_cmd(args []string) error {
	// run the command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func logDebug(format string, args ...interface{}) {
	if os.Getenv("CRONLOCK_DEBUG") == "true" {
		date := time.Now().Format(logDateFormat)
		fmt.Printf(date+" DEBUG: "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	date := time.Now().Format(logDateFormat)
	fmt.Printf(date+" INFO: "+format+"\n", args...)
}
