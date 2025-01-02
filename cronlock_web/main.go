// websever to manage cronlock
// requests currenlty running commands and lists them in a table showing hash, status, duration, and command
package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	redis "github.com/redis/go-redis/v9"
)

var (
	CRONLOCK_HOST = getEnvStr("CRONLOCK_HOST", "localhost")
	CRONLOCK_PORT = getEnvInt("CRONLOCK_PORT", 6379)
)

func main() {
	// create a new http server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get the keys with 'status' value of "RUNNING" in Redis
		redisClient := connect()
		defer redisClient.Close()

		// get the keys with 'status' value of "RUNNING" in Redis
		keys, err := redisClient.Keys(context.Background(), "*").Result()
		if err != nil {
			log.Fatalf("Failed to get keys: %v\n", err)
		}

		type Row struct {
			Hash     string
			Status   string
			Duration string
			Command  string
		}

		var rows []Row
		for _, key := range keys {
			if key != "example" {
				log.Printf("Key: %s\n", key)
				value, err := redisClient.Get(context.Background(), key).Result()
				if err != nil {
					log.Fatalf("Failed to get value for key %s: %v\n", key, err)
				}

				rows = append(rows, Row{Hash: key, Status: value, Duration: "duration", Command: "command"})
			}
		}

		tmpl := template.Must(template.New("table").Parse(`
			<html>
			<body>
				<table>
					<tr>
						<th>Hash</th>
						<th>Status</th>
						<th>Duration</th>
						<th>Command</th>
					</tr>
					{{range .}}
					<tr>
						<td>{{.Hash}}</td>
						<td>{{.Status}}</td>
						<td>{{.Duration}}</td>
						<td>{{.Command}}</td>
					</tr>
					{{end}}
				</table>
			</body>
			</html>
		`))

		if err := tmpl.Execute(w, rows); err != nil {
			log.Fatalf("Failed to execute template: %v\n", err)
		}
	})
	http.ListenAndServe(":8080", nil)
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
