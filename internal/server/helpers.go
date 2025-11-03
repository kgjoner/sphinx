package server

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

/* =====================================================================
	Setup
===================================================================== */

func realIP() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ip string

			if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
				ip = cfIP
			} else if tcIP := r.Header.Get("True-Client-IP"); tcIP != "" {
				ip = tcIP
			} else if xrIP := r.Header.Get("X-Real-IP"); xrIP != "" {
				ip = xrIP
			} else if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				i := strings.Index(xff, ",")
				if i == -1 {
					i = len(xff)
				}
				ip = xff[:i]
			}

			if ip == "" || net.ParseIP(ip) == nil {
				next.ServeHTTP(w, r)
			}
			
			r.RemoteAddr = ip
			next.ServeHTTP(w, r)
		})
	}
}

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
