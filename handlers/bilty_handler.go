package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"hariomtransport/models"
	"hariomtransport/repository"
)

type BiltyHandler struct {
	Repo repository.BiltyRepository
}

// CreateBilty handler
func (h *BiltyHandler) CreateBilty(w http.ResponseWriter, r *http.Request) {
	var bilty models.Bilty
	if err := json.NewDecoder(r.Body).Decode(&bilty); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Repo.CreateBiltyWithParties(&bilty); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(bilty)
}

// GetAllBilty handler
func (h *BiltyHandler) GetAllBilty(w http.ResponseWriter, r *http.Request) {
	filters := make(map[string]interface{})
	q := r.URL.Query()
	for key, values := range q {
		if len(values) > 0 && values[0] != "" {
			// Attempt to convert numeric values to int if possible
			if intVal, err := strconv.Atoi(values[0]); err == nil {
				filters[key] = intVal
			} else {
				filters[key] = values[0]
			}
		}
	}

	list, err := h.Repo.GetBilty(filters, false) // fetch multiple
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if list == nil {
		list = []*models.Bilty{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// GetBiltyByID handler
func (h *BiltyHandler) GetBiltyByID(w http.ResponseWriter, r *http.Request, id string) {
	// Assuming ID is int64
	biltyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "invalid bilty ID", http.StatusBadRequest)
		return
	}

	filters := map[string]interface{}{"id": biltyID}
	list, err := h.Repo.GetBilty(filters, true) // fetch single
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(list) == 0 {
		http.Error(w, "Bilty not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list[0])
}

func (h *BiltyHandler) DeleteBilty(w http.ResponseWriter, r *http.Request) {

	biltyIDStr := r.URL.Query().Get("id")
	if biltyIDStr == "" {
		http.Error(w, "missing bilty id", http.StatusBadRequest)
		return
	}

	biltyID, err := strconv.ParseInt(biltyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid bilty id", http.StatusBadRequest)
		return
	}

	// Attempt to delete bilty
	if err := h.Repo.DeleteBilty(biltyID); err != nil {
		http.Error(w, "failed to delete bilty: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"success":true,"message":"Bilty deleted successfully"}`))
}
