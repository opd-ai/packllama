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
			bodyBytes := captureRequestBody(r, logRequests)
			recorder := newStatusRecorder(w, logResponses)
			start := time.Now()
			next.ServeHTTP(recorder, r)

			args := baseLogArgs(r, recorder, time.Since(start))
			args = appendVerboseLogs(args, bodyBytes, recorder, logRequests, logResponses)
			logger.Log(r.Context(), httpLogLevel(recorder.statusCode), "request completed", args...)
		})
	}
}

// captureRequestBody reads and restores the request body when logRequests is true.
func captureRequestBody(r *http.Request, enabled bool) []byte {
	if !enabled || r.Body == nil {
		return nil
	}
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	return b
}

// newStatusRecorder wraps w and optionally enables response body capture.
func newStatusRecorder(w http.ResponseWriter, captureBody bool) *statusRecorder {
	rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
	if captureBody {
		rec.buf = &bytes.Buffer{}
	}
	return rec
}

// baseLogArgs returns the fixed set of log fields present on every request.
func baseLogArgs(r *http.Request, rec *statusRecorder, dur time.Duration) []any {
	return []any{
		"method", r.Method,
		"path", r.URL.Path,
		"status", rec.statusCode,
		"duration", dur,
		"bytes", rec.bytesWritten,
		"request_id", RequestIDFromContext(r.Context()),
	}
}

// appendVerboseLogs optionally appends request/response body fields.
func appendVerboseLogs(args []any, reqBody []byte, rec *statusRecorder, logReq, logResp bool) []any {
	if logReq && len(reqBody) > 0 {
		args = append(args, "request_body", string(reqBody))
	}
	if logResp && rec.buf != nil && rec.buf.Len() > 0 {
		args = append(args, "response_body", rec.buf.String())
	}
	return args
}

// httpLogLevel maps HTTP status codes to slog levels.
func httpLogLevel(statusCode int) slog.Level {
	switch {
	case statusCode >= 500:
		return slog.LevelError
	case statusCode >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
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
