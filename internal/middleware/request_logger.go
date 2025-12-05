package middleware

import (
	"net/http"
	"time"

	"fire-go/internal/logger"
)

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		// ResponseWriter modificado para capturar status
		sw := &statusResponseWriter{
			ResponseWriter: w,
			status:         200,
		}

		next.ServeHTTP(sw, r)

		duration := time.Since(start)

		logger.Log.Info("http_request",
			"path", r.URL.Path,
			"method", r.Method,
			"status", sw.status,
			"duration_ms", duration.Milliseconds(),
		)
	})
}
