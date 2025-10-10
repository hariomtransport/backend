package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hariomtransport/models"
	"github.com/hariomtransport/repository"
)

type InitialHandler struct {
	Repo repository.InitialRepository
}

func (h *InitialHandler) SaveInitial(w http.ResponseWriter, r *http.Request) {
	var initial models.InitialSetup
	if err := json.NewDecoder(r.Body).Decode(&initial); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Repo.SaveInitial(&initial); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(initial)
}

func (h *InitialHandler) GetInitial(w http.ResponseWriter, r *http.Request) {
	initial, err := h.Repo.GetInitial()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if initial == nil {
		http.Error(w, "Initial details not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(initial)
}
