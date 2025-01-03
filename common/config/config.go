package config

import (
	"os"
	"strconv"
)

var (
	CRONLOCK_HOST           = getEnvStr("CRONLOCK_HOST", "localhost")
	CRONLOCK_PORT           = getEnvStr("CRONLOCK_PORT", "6379")
	CRONLOCK_RETRY_ATTEMPTS = getEnvInt("CRONLOCK_RETRY_ATTEMPTS", 5)
	CRONLOCK_DEBUG          = getEnvStr("CRONLOCK_DEBUG", "false")
	CRONLOCK_KEEP_HISTORY   = getEnvStr("CRONLOCK_KEEP_HISTORY", "false")
	CRONLOCK_EXPIRY_TIME    = getEnvInt("CRONLOCK_EXPIRY_TIME", 86400)

	CRONWEB_HOST = getEnvStr("CRONWEB_HOST", "localhost")
	CRONWEB_PORT = getEnvStr("CRONWEB_PORT", "8080")
)

const (
	CRON_STATUS_RUNNING  = "RUNNING"
	CRON_STATUS_SUCCESS  = "SUCCESS"
	CRON_STATUS_FAILED   = "FAILED"
	CRON_STATUS_COMPLETE = "COMPLETE"
)

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
