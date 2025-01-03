package cron

import (
	"testing"

	"github.com/devinodaniel/cronlock-go/common/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisFailure(t *testing.T) {
	args := []string{"echo", "hello"}
	cron := New(args)
	cron.RedisClient = nil

	err := cron.Run()
	assert.Error(t, err)
	assert.Equal(t, "redis client not set", err.Error())
}

func TestNew(t *testing.T) {
	args := []string{"echo", "hello"}
	cron := New(args)

	assert.NotNil(t, cron)
	assert.Equal(t, args, cron.Args)
	assert.NotEmpty(t, cron.Md5Hash)
}

func TestMarshalUnmarshalBinary(t *testing.T) {
	args := []string{"echo", "hello"}
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

func TestRunSuccess(t *testing.T) {
	// Mock Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Mock config values
	config.CRONLOCK_EXPIRY_TIME = 5
	config.CRONLOCK_KEEP_HISTORY = "false"

	args := []string{"sleep", "5"}
	cron := New(args)
	cron.RedisClient = client

	err := cron.Run()
	assert.NoError(t, err)
	assert.Equal(t, config.CRON_STATUS_COMPLETE, cron.Status)
	assert.NotEmpty(t, cron.EpochStart)
	assert.NotEmpty(t, cron.EpochEnd)
	assert.Equal(t, int64(5), cron.Duration)
}

func TestRunFail(t *testing.T) {
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

func TestRunInvalid(t *testing.T) {
	// Mock Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Mock config values
	config.CRONLOCK_EXPIRY_TIME = 1

	args := []string{"not", "a", "command"}
	cron := New(args)
	cron.RedisClient = client

	err := cron.Run()
	assert.NoError(t, err)
	assert.Equal(t, config.CRON_STATUS_FAILED, cron.Status)
	assert.NotEmpty(t, cron.EpochStart)
	assert.NotEmpty(t, cron.EpochEnd)
	assert.Equal(t, int64(0), cron.Duration)
	assert.Equal(t, "exec: \"not\": executable file not found in $PATH", cron.Error)
}

func TestHash(t *testing.T) {
	args := []string{"echo", "hello"}
	hash1 := hash(args)
	hash2 := hash(args)

	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}
