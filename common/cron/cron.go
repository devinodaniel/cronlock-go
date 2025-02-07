package cron

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/devinodaniel/cronlock-go/common/config"
	"github.com/devinodaniel/cronlock-go/common/log"

	"github.com/redis/go-redis/v9"
)

type Cron struct {
	RedisClient *redis.Client
	Ctx         context.Context `json:"-"`
	EpochStart  int64           `json:"epochStart"`
	EpochEnd    int64           `json:"epochEnd"`
	Timeout     time.Duration   `json:"timeout"`
	ExpiryTime  time.Duration   `json:"expiryTime"`
	Duration    int64           `json:"duration"`
	Md5Hash     string          `json:"md5Hash"`
	Args        []string        `json:"args"`
	Status      string          `json:"status"`
	Pid         int             `json:"pid"`
	Error       string          `json:"error"`
}

func New(args []string) *Cron {
	md5hash := hash(args)

	if config.CRONLOCK_PRINT_ARGS {
		log.Info("args: %v", args)
	}

	return &Cron{
		Ctx:     context.Background(),
		Args:    args,
		Md5Hash: md5hash,
	}
}

// MarshalBinary automatically marshals the cron object into a JSON object to be stored in Redis
func (c Cron) MarshalBinary() ([]byte, error) {
	// Custom logic to marshal your type into JSON
	bytes, err := json.Marshal(c)
	return bytes, err
}

// UnmarshalBinary automatically unmarshals the cron object from a JSON object stored in Redis
func (c *Cron) UnmarshalBinary(data []byte) error {
	// Custom logic to unmarshal your type from JSON
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	return nil
}

func (c *Cron) Run() error {
	if c.RedisClient == nil {
		return fmt.Errorf("redis client not set")
	}

	// run the command and set the metadata
	if err := c.start(); err != nil {
		return err
	}

	if err := c.finish(); err != nil {
		return err
	}

	return nil
}

// start() runs the command and sets the metadata if the command is not already running
func (c *Cron) start() error {
	// set the start time
	c.EpochStart = time.Now().Unix()
	// set the status to running
	c.Status = config.CRON_STATUS_RUNNING
	// expire and auto-unlock after N amount of time in case the command fails to finish
	c.Timeout = time.Duration(config.CRONLOCK_TIMEOUT) * time.Second

	// check if the command is already running by setting the key
	// SetNX returns true if the key was set, false if it already exists
	okToRun, err := c.RedisClient.SetNX(c.Ctx, c.Md5Hash, c, c.Timeout).Result()
	if err != nil {
		return err
	}

	// bail early if the key was already set, the command is already running
	if !okToRun {
		c.Status = config.CRON_STATUS_SKIPPED
		return fmt.Errorf("%s locked, skipping", c.Md5Hash)
	}

	// if the key was set, the command is ok to run
	if okToRun {
		log.Debug("%s added to Redis", c.Md5Hash)
		log.Info("%s started, locking", c.Md5Hash)

		// run the command, get the Process ID
		if err := raw_cmd(c.Args); err != nil {
			c.Error = err.Error()
			c.Status = config.CRON_STATUS_FAILED
			// command failures are okay, we just want to know if it failed
			return nil
		}

		c.Status = config.CRON_STATUS_SUCCESS
	}

	return nil
}

// finish() updates the metadata after the command has executed
func (c *Cron) finish() error {
	// set the end time
	c.EpochEnd = time.Now().Unix()
	// calculate the duration
	c.Duration = c.EpochEnd - c.EpochStart

	// set the expiry time
	if c.Timeout == 0 {
		// keep the lock
		c.ExpiryTime = c.Timeout
	} else {
		// unlock after N amount of time
		c.ExpiryTime = time.Duration(config.CRONLOCK_GRACE_PERIOD) * time.Second
	}

	// set the status to complete if it was successful
	// other statuses should remain the same (failed, skipped, etc)
	if c.Status == config.CRON_STATUS_SUCCESS {
		c.Status = config.CRON_STATUS_COMPLETE
	}

	// update the metadata in Redis
	updateResult, err := c.RedisClient.Set(c.Ctx, c.Md5Hash, c, c.ExpiryTime).Result()
	if err != nil {
		return err
	}
	if updateResult == "OK" {
		log.Debug("%s %s metadata updated in Redis", c.Md5Hash, c.Status)
	}

	log.Info("%s finished in %d sec, unlocking in %v", c.Md5Hash, c.Duration, c.ExpiryTime)

	return nil
}

// raw_cmd() executes the cron commands or scripts and returns an error if it fails
func raw_cmd(args []string) error {
	// run the command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if config.CRONLOCK_PRINT_STDOUT {
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// hash creates an md5 hash of all the arugments passed to it
func hash(args ...interface{}) string {
	// return the md5 hash of the arguments
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprint(args...))))
}
