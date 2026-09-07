package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(statusCode int) {
	rec.statusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &responseRecorder{w, http.StatusOK}
		next.ServeHTTP(rec, r)
		duration := time.Since(start)

		zap.L().Info("HTTP Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rec.statusCode),
			zap.Duration("duration", duration), // Zap punya tipe khusus untuk durasi!
			zap.String("ip", r.RemoteAddr),
		)
	})
}
