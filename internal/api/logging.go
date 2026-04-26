package api

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nicopiov/zerapi/internal/util"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return w.ResponseWriter.Write(body)
}

func WithLogging(next http.Handler, output io.Writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		writer := &statusWriter{
			ResponseWriter: w,
		}

		next.ServeHTTP(writer, r)

		status := writer.status
		if status == 0 {
			status = http.StatusOK
		}

		fmt.Fprintf(
			output,
			"%-6s %-24s %s %s\n",
			r.Method,
			r.URL.Path,
			util.Status(status),
			formatDuration(time.Since(start)),
		)
	})
}

func formatDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return "0ms"
	}

	return duration.Round(time.Millisecond).String()
}
