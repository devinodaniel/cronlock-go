package config

import (
	"log"
	"os"
	"strconv"
)

var (
	CRONLOCK_REDIS_HOST     = EnvStr("CRONLOCK_REDIS_HOST", "localhost")
	CRONLOCK_REDIS_PORT     = EnvStr("CRONLOCK_REDIS_PORT", "6379")
	CRONLOCK_REDIS_DATABASE = EnvInt("CRONLOCK_REDIS_DATABASE", 0)
	CRONLOCK_TIMEOUT        = EnvInt("CRONLOCK_TIMEOUT", 3600)
	CRONLOCK_GRACE_PERIOD   = EnvInt("CRONLOCK_GRACE_PERIOD", 5)

	// useful for debugging
	CRONLOCK_DEBUG        = EnvBool("CRONLOCK_DEBUG", false)
	CRONLOCK_PRINT_STDOUT = EnvBool("CRONLOCK_PRINT_STDOUT", false)
	CRONLOCK_PRINT_ARGS   = EnvBool("CRONLOCK_PRINT_ARGS", false)

	// web server
	CRONWEB_HOST = EnvStr("CRONWEB_HOST", "localhost")
	CRONWEB_PORT = EnvStr("CRONWEB_PORT", "8080")
)

const (
	CRON_STATUS_RUNNING  = "RUNNING"
	CRON_STATUS_SUCCESS  = "SUCCESS"
	CRON_STATUS_FAILED   = "FAILED"
	CRON_STATUS_COMPLETE = "COMPLETE"
	CRON_STATUS_SKIPPED  = "SKIPPED"
)

// EnvString retrieves the string value of the environment variable named by the key.
func EnvStr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	} else {
		return defaultValue
	}

}

// EnvInt retrieves the integer value of the environment variable named by the key.
func EnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		// make sure that the value is an integer
		value, _ := strconv.Atoi(value)
		return value
	} else {
		return defaultValue
	}
}

// EnvBool retrieves the boolean value of the environment variable named by the key.
func EnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true":
			return true
		case "false":
			return false
		default:
			log.Fatalf("Invalid boolean value for %s: %s", key, value)
		}
	}
	return defaultValue
}
