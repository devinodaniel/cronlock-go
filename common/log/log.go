package log

import (
	"fmt"
	"log"
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
	log.Fatalf(date+" FATAL: "+format+"\n", args...)
}

func Error(format string, args ...interface{}) {
	date := time.Now().Format(logDateFormat)
	log.Fatalf(date+" ERROR: "+format+"\n", args...)
}
