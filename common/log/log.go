package log

import (
	"fmt"
	"os"
	"time"
)

var (
	logDateFormat = "2006-01-02 15:04:05"
)

func Debug(format string, args ...interface{}) {
	if os.Getenv("CRONLOCK_DEBUG") == "true" {
		date := time.Now().Format(logDateFormat)
		fmt.Printf(date+" DEBUG: "+format+"\n", args...)
	}
}

func Info(format string, args ...interface{}) {
	date := time.Now().Format(logDateFormat)
	fmt.Printf(date+" INFO: "+format+"\n", args...)
}

func Fatal(format string, args ...interface{}) {
	date := time.Now().Format(logDateFormat)
	fmt.Printf(date+" FATAL: "+format+"\n", args...)
	os.Exit(1)
}

func Error(format string, args ...interface{}) {
	date := time.Now().Format(logDateFormat)
	fmt.Printf(date+" ERROR: "+format+"\n", args...)
}
