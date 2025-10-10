package handlers

import (
	"fmt"
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
				fmt.Printf("\n=== PANIC RECOVERED ===\nError: %v\nStacktrace:\n%s\n===================================\n", rec, string(stack))
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()

		handler(w, r)
	}
}
