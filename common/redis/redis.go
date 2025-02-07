package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Connect() makes a connection to redis and retries if it fails
func Connect(host string, port string, database int) (*redis.Client, error) {
	// connect to redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: "",       // No password set
		DB:       database, // Default is 0, 1 is for tests
		Protocol: 2,        // Connection protocol
	})

	// ping redis to check if the connection is working
	pong, err := client.Ping(context.Background()).Result()
	if pong != "PONG" || err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	return client, nil
}
