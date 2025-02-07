package redis

import (
	"fmt"
	"os"
	"testing"

	"github.com/devinodaniel/cronlock-go/common/config"
	"github.com/stretchr/testify/assert"
)

func TestConnectRedisSuccess(t *testing.T) {
	client, err := Connect(config.CRONLOCK_REDIS_HOST, config.CRONLOCK_REDIS_PORT, 1)

	if err != nil {
		fmt.Printf("\n*********FAILED TO CONNECT TO REDIS - IS THE LOCAL REDIS SERVER RUNNING?\n"+
			"%v\n"+
			"This test should always pass. It is REQUIRED for other tests to pass.\n\n", err)
		os.Exit(1)
	}

	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestConnectRedisFailure(t *testing.T) {
	client, err := Connect("localhost", "1234", 1)
	assert.Error(t, err)
	assert.Nil(t, client)
}
