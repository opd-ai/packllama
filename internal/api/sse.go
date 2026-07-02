package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// writeSSEEvent marshals payload as a Server-Sent Events "data:" line and flushes.
func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// writeSSEError emits an SSE error event and flushes.
func writeSSEError(w http.ResponseWriter, flusher http.Flusher, err error) {
	fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
	flusher.Flush()
}
