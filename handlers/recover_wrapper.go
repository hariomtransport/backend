package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
)

// RecoverWrapper wraps an http.HandlerFunc with panic recovery
func RecoverWrapper(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := make([]byte, 8*1024)
				stack = stack[:runtime.Stack(stack, false)]

				// Log detailed panic info to server logs
				log.Printf(`
				=== PANIC RECOVERED ===
				Error: %v
				Request: %s %s
				Stacktrace:
				%s
				===================================
				`, rec, r.Method, r.URL.Path, string(stack))

				// Send consistent ApiResponse to client
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(ApiResponse{
					Success: false,
					Message: "Internal server error",
				})
			}
		}()

		handler(w, r)
	}
}
