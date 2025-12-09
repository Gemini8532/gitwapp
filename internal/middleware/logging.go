package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware provides structured logging for HTTP requests using the slog package.
// It logs the method, URL, status code, duration, remote address, and user agent
// for each request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer that captures the status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(lrw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the request details
		slog.Info("http_request",
			"method", r.Method,
			"url", r.URL.Path,
			"status", lrw.statusCode,
			"duration", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

// loggingResponseWriter is a wrapper around http.ResponseWriter that captures the
// status code of the response.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code and calls the underlying ResponseWriter's
// WriteHeader method.
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
