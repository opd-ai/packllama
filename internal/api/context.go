package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

type requestIDKey struct{}

func withRequestIDContext(ctx context.Context) (context.Context, string) {
	requestID := newRequestID()
	return context.WithValue(ctx, requestIDKey{}, requestID), requestID
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err == nil {
		return hex.EncodeToString(buf)
	}
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
