package api

import (
	"net/http"
	"time"
)

func WithDelay(next http.Handler, delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if delay > 0 {
			time.Sleep(delay)
		}

		next.ServeHTTP(w, r)
	})
}
