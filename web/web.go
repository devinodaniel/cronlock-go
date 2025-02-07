// websever to manage cronlock
// requests currenlty running commands and lists them in a table showing hash, status, duration, and command
package web

import (
	"context"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/devinodaniel/cronlock-go/common/config"
	"github.com/devinodaniel/cronlock-go/common/cron"
	"github.com/devinodaniel/cronlock-go/common/log"
	"github.com/devinodaniel/cronlock-go/common/redis"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	humanDateFormat = "2006-01-02 15:04:05"
)

type CronRow struct {
	Hash     string
	Status   string
	Started  string
	Duration int64
	Command  string
	Error    string
}

func Server() {
	http.HandleFunc("/", listCrons)
	// prometheus metrics
	http.Handle("/metrics", promhttp.Handler())

	log.Info("Starting server on :%s", config.CRONWEB_PORT)
	http.ListenAndServe(":"+config.CRONWEB_PORT, nil)
}

// prometheus Handler for metrics
func prometheusHandler(w http.ResponseWriter, req *http.Request) {
	promhttp.Handler().ServeHTTP(w, req)
}

func listCrons(w http.ResponseWriter, req *http.Request) {
	// open redis connection
	redisClient, err := redis.Connect(config.CRONLOCK_REDIS_HOST, config.CRONLOCK_REDIS_PORT, config.CRONLOCK_REDIS_DATABASE)
	if err != nil {
		log.Error("Failed to connect to Redis: %v\n", err)
	}
	defer redisClient.Close()

	// get all keys from redis
	keys, err := redisClient.Keys(context.Background(), "*").Result()
	if err != nil {
		log.Error("Failed to get keys: %v\n", err)
	}

	var rows []CronRow

	// iterate over all keys and get the values
	for _, key := range keys {
		if key != "example" {
			value, err := redisClient.Get(context.Background(), key).Result()
			if err != nil {
				log.Error("Failed to get value for key %s: %v\n", key, err)
			}

			cron := cron.New([]string{})
			if err := cron.UnmarshalBinary([]byte(value)); err != nil {
				log.Error("Failed to unmarshal value for key %s: %v\n", key, err)
			}

			// choose which duration to display based on the status of the cron job
			var duration int64
			switch cron.Status {
			// if the cron job is running, calculate the duration based on the start time
			case config.CRON_STATUS_RUNNING:
				duration = time.Now().Unix() - cron.EpochStart
			// if the cron job is complete, finished duration from the metadata
			case config.CRON_STATUS_COMPLETE, config.CRON_STATUS_FAILED:
				duration = cron.Duration
			}

			humanStartTime := time.Unix(cron.EpochStart, 0).Format(humanDateFormat)

			rows = append(rows,
				CronRow{
					Command:  strings.Join(cron.Args, " "),
					Status:   cron.Status,
					Started:  humanStartTime,
					Duration: duration,
					Hash:     cron.Md5Hash,
					Error:    cron.Error})
		}
	}

	tmp := template.Must(template.ParseFiles("templates/list_crons.html"))
	tmp.Execute(w, rows)
}
