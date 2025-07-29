package server

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

/* =====================================================================
	Metrics
===================================================================== */

var (
	RequestCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_request_count",
		Help: "The total number of requests",
	})
)

func countRequestMetric() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			RequestCounter.Inc()
			next.ServeHTTP(w, r)
		})
	}
}

/* =====================================================================
	Cron Jobs
===================================================================== */

// Run a task in every defined interval.
// The task function will be executed periodically until the context is closed.
//
// The interval is defined in seconds.
func runPeriodicTask(
	ctx context.Context,
	interval int,
	task func(),
) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for {
		select {
		// The ticker sends a time data every period
		case <-ticker.C:
			task()

		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
