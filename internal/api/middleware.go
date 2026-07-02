package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		var requestID string
		if incoming := r.Header.Get("X-Request-ID"); incoming != "" {
			requestID = incoming
			ctx = context.WithValue(r.Context(), requestIDKey{}, requestID)
		} else {
			ctx, requestID = withRequestIDContext(r.Context())
		}
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withLogging(logger *slog.Logger, logRequests, logResponses bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var bodyBytes []byte
			if logRequests && r.Body != nil {
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}

			recorder := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			if logResponses {
				recorder.buf = &bytes.Buffer{}
			}
			start := time.Now()
			next.ServeHTTP(recorder, r)

			args := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusCode,
				"duration", time.Since(start),
				"bytes", recorder.bytesWritten,
				"request_id", RequestIDFromContext(r.Context()),
			}
			if logRequests && len(bodyBytes) > 0 {
				args = append(args, "request_body", string(bodyBytes))
			}
			if logResponses && recorder.buf != nil && recorder.buf.Len() > 0 {
				args = append(args, "response_body", recorder.buf.String())
			}

			level := slog.LevelInfo
			if recorder.statusCode >= 500 {
				level = slog.LevelError
			} else if recorder.statusCode >= 400 {
				level = slog.LevelWarn
			}
			logger.Log(r.Context(), level, "request completed", args...)
		})
	}
}

func withRecovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("request panicked",
						"error", fmt.Sprint(recovered),
						"request_id", RequestIDFromContext(r.Context()),
					)
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func withCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowsOrigin(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func allowsOrigin(allowedOrigins []string, origin string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || strings.EqualFold(allowedOrigin, origin) {
			return true
		}
	}
	return false
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
	buf          *bytes.Buffer // non-nil when response body capture is enabled
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytesWritten += int64(n)
	if r.buf != nil {
		r.buf.Write(b[:n])
	}
	return n, err
}

// Flush implements http.Flusher by forwarding to the underlying writer when supported.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
