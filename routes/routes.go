package routes

import (
	"net/http"

	"github.com/hariomtransport/handlers"
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
	pdfHandler *handlers.PDFHandler,
) {
	// User routes
	http.Handle("/signup", withCORS(http.HandlerFunc(handlers.RecoverWrapper(userHandler.Signup))))
	http.Handle("/login", withCORS(http.HandlerFunc(handlers.RecoverWrapper(userHandler.Login))))
	http.Handle("/bilty/pdf", withCORS(http.HandlerFunc(handlers.RecoverWrapper(pdfHandler.BiltyPDF))))

	// Bilty routes
	http.Handle("/bilty", withCORS(http.HandlerFunc(handlers.RecoverWrapper(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			biltyHandler.CreateBilty(w, r)
		case http.MethodGet:
			biltyHandler.GetAllBilty(w, r)
		case http.MethodDelete:
			biltyHandler.DeleteBilty(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))))

	// Get bilty by ID
	http.Handle("/bilty/", withCORS(http.HandlerFunc(handlers.RecoverWrapper(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/bilty/"):]
		if id != "" {
			biltyHandler.GetBiltyByID(w, r, id)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))))

	// Initial setup routes
	http.Handle("/initial", withCORS(http.HandlerFunc(handlers.RecoverWrapper(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			initialHandler.SaveInitial(w, r)
		case http.MethodGet:
			initialHandler.GetInitial(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))))
}
