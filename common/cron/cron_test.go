package cron

import (
	"fmt"
	"testing"
	"time"

	"github.com/devinodaniel/cronlock-go/common/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisClientNotSet(t *testing.T) {
	args := []string{"echo", "hello 1"}
	cron := New(args)
	cron.RedisClient = nil

	err := cron.Run()
	assert.Error(t, err)
	assert.Equal(t, "redis client not set", err.Error())
}

func TestNewCron(t *testing.T) {
	args := []string{"echo", "hello 2"}
	cron := New(args)

	assert.NotNil(t, cron)
	assert.Equal(t, args, cron.Args)
	assert.NotEmpty(t, cron.Md5Hash)
}

func TestCronMarshalUnmarshalBinary(t *testing.T) {
	args := []string{"echo", "hello 3"}
	cron := New(args)

	data, err := cron.MarshalBinary()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	var newCron Cron
	err = newCron.UnmarshalBinary(data)
	assert.NoError(t, err)
	assert.Equal(t, cron.Args, newCron.Args)
	assert.Equal(t, cron.Md5Hash, newCron.Md5Hash)
}

func TestCronRunSuccess(t *testing.T) {
	// Mock Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	args := []string{"sleep", "2"}
	cron := New(args)
	cron.RedisClient = client

	err := cron.Run()
	assert.NoError(t, err)
	assert.Equal(t, config.CRON_STATUS_COMPLETE, cron.Status)
	assert.NotEmpty(t, cron.EpochStart)
	assert.NotEmpty(t, cron.EpochEnd)
	assert.Equal(t, int64(2), cron.Duration)
}

func TestCronRunSkipped(t *testing.T) {
	// Mock Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// two cron jobs with the same arguments
	// the second one should be skipped becasue the first one is still running
	args1 := []string{"sleep", "5"}
	cron1 := New(args1)
	cron1.RedisClient = client
	go cron1.Run()

	// Since Sleep to allow cron1 to start in order to lock for cron2
	time.Sleep(2 * time.Second)

	args2 := []string{"sleep", "5"}
	cron2 := New(args2)
	cron2.RedisClient = client
	cron2.Run()

	sleepFinish := time.Duration(6)
	// Sleep to allow cron1 to finish in order to get the status
	time.Sleep(sleepFinish * time.Second)

	fmt.Printf("cron1 status after %d seconds: %s\n", sleepFinish, cron1.Status)

	// cron1
	assert.Equal(t, config.CRON_STATUS_COMPLETE, cron1.Status)
	assert.NotEmpty(t, cron1.EpochStart)
	assert.NotEmpty(t, cron1.EpochEnd)
	assert.Equal(t, int64(5), cron1.Duration)

	// cron2
	assert.Equal(t, config.CRON_STATUS_SKIPPED, cron2.Status)
	assert.NotEmpty(t, cron2.EpochStart)
	// cron2 should not have an end time, because it was skipped
	assert.Empty(t, cron2.EpochEnd)
}

func TestCronRunCommandExitCode1(t *testing.T) {
	// Mock Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Mock config values
	config.CRONLOCK_EXPIRY_TIME = 1

	args := []string{"test", "-f", "/tmp/doesnotexist"}
	cron := New(args)
	cron.RedisClient = client

	err := cron.Run()
	assert.NoError(t, err)
	assert.Equal(t, config.CRON_STATUS_FAILED, cron.Status)
	assert.NotEmpty(t, cron.EpochStart)
	assert.NotEmpty(t, cron.EpochEnd)
	assert.Equal(t, "exit status 1", cron.Error)
}

func TestCronRunInvalid(t *testing.T) {
	// Mock Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Mock config values
	config.CRONLOCK_EXPIRY_TIME = 1

	args := []string{"lkjhlaskjdft", "not", "an", "executable"}
	cron := New(args)
	cron.RedisClient = client

	err := cron.Run()
	assert.NoError(t, err)
	assert.Equal(t, config.CRON_STATUS_FAILED, cron.Status)
	assert.NotEmpty(t, cron.EpochStart)
	assert.NotEmpty(t, cron.EpochEnd)
	assert.Equal(t, int64(0), cron.Duration)
	assert.Equal(t, "exec: \"lkjhlaskjdft\": executable file not found in $PATH", cron.Error)
}

func TestCronMd5Hash(t *testing.T) {
	args := []string{"echo", "hello 4"}
	arg_md5hash := hash(args)
	real_md5hash := "fbe88f94751af8c5e4e2dc7fbd461b08"

	assert.Equal(t, arg_md5hash, real_md5hash)
}
