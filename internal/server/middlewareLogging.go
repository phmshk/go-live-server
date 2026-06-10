package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if pathShouldBeIgnored(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		startTime := time.Now()

		rww := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(rww, r)

		duration := time.Since(startTime)

		var durationStr string
		if duration < time.Millisecond {
			durationStr = fmt.Sprintf("%dµs", duration.Microseconds())
		} else {
			durationStr = fmt.Sprintf("%.1fms", float64(duration.Nanoseconds())/1e6)
		}

		log.Printf(
			"[HTTP] %-7s %-40s | Status: %3d | %10s",
			r.Method,
			r.URL.Path,
			rww.statusCode,
			durationStr,
		)
	})
}

func pathShouldBeIgnored(path string) bool {
	if path == "/live-reload" {
		return true
	}

	if strings.Contains(path, "/.well-known") {
		return true
	}

	return false
}
