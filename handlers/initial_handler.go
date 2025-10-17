package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hariomtransport/backend/models"
	"github.com/hariomtransport/backend/repository"
)

type InitialHandler struct {
	Repo repository.InitialRepository
}

// SaveInitial handler
func (h *InitialHandler) SaveInitial(w http.ResponseWriter, r *http.Request) {
	var initial models.InitialSetup
	if err := json.NewDecoder(r.Body).Decode(&initial); err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.Repo.SaveInitial(&initial); err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to save initial setup: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Initial setup saved successfully",
		Data:    initial,
	})
}

// GetInitial handler
func (h *InitialHandler) GetInitial(w http.ResponseWriter, r *http.Request) {
	initial, err := h.Repo.GetInitial()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to fetch initial setup: " + err.Error(),
		})
		return
	}

	if initial == nil {
		writeJSON(w, http.StatusNotFound, ApiResponse{
			Success: false,
			Message: "Initial setup not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Initial setup fetched successfully",
		Data:    initial,
	})
}
