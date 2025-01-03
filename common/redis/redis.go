package redis

import (
	"context"
	"fmt"

	"github.com/devinodaniel/cronlock-go/common/config"

	"github.com/redis/go-redis/v9"
)

// Connect() makes a connection to redis and retries if it fails
func Connect() (*redis.Client, error) {
	// connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.CRONLOCK_HOST, config.CRONLOCK_PORT),
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

	// ping redis to check if the connection is working
	pong, err := client.Ping(context.Background()).Result()
	if pong != "PONG" || err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	return client, nil
}
