package routes

import (
	"net/http"

	"hariomtransport/handlers"
)

// CORS middleware
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace * with your domain in production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func SetupRoutes(
	userHandler *handlers.UserHandler,
	biltyHandler *handlers.BiltyHandler,
	initialHandler *handlers.InitialHandler,
) {
	// User routes
	http.Handle("/signup", withCORS(http.HandlerFunc(userHandler.Signup)))
	http.Handle("/login", withCORS(http.HandlerFunc(userHandler.Login)))

	// Bilty routes
	http.Handle("/bilty", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			biltyHandler.CreateBilty(w, r)
		case http.MethodGet:
			biltyHandler.GetAllBilty(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))

	// Get bilty by ID
	http.Handle("/bilty/", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/bilty/"):]
		if id != "" {
			biltyHandler.GetBiltyByID(w, r, id)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})))

	// Initial setup routes
	http.Handle("/initial", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			initialHandler.SaveInitial(w, r)
		case http.MethodGet:
			initialHandler.GetInitial(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})))
}
