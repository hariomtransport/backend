package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/hariomtransport/backend/models"
	"github.com/hariomtransport/backend/repository"
)

// Response structure for consistent API responses
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type BiltyHandler struct {
	Repo repository.BiltyRepository
}

// Helper to write JSON responses
func writeJSON(w http.ResponseWriter, status int, resp ApiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

// CreateBilty handler
func (h *BiltyHandler) CreateBilty(w http.ResponseWriter, r *http.Request) {
	var bilty models.Bilty
	if err := json.NewDecoder(r.Body).Decode(&bilty); err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.Repo.CreateBiltyWithParties(&bilty); err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to create bilty: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "Bilty created successfully",
		Data:    bilty,
	})
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

	list, err := h.Repo.GetBilty(filters, false)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to fetch bilty records: " + err.Error(),
		})
		return
	}

	if list == nil {
		list = []*models.Bilty{}
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Bilty records fetched successfully",
		Data:    list,
	})
}

// GetBiltyByID handler
func (h *BiltyHandler) GetBiltyByID(w http.ResponseWriter, r *http.Request, id string) {
	biltyID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid bilty ID",
		})
		return
	}

	filters := map[string]interface{}{"id": biltyID}
	list, err := h.Repo.GetBilty(filters, true)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to fetch bilty: " + err.Error(),
		})
		return
	}
	if len(list) == 0 {
		writeJSON(w, http.StatusNotFound, ApiResponse{
			Success: false,
			Message: "Bilty not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Bilty details fetched successfully",
		Data:    list[0],
	})
}

// DeleteBilty handler
func (h *BiltyHandler) DeleteBilty(w http.ResponseWriter, r *http.Request) {
	biltyIDStr := r.URL.Query().Get("id")
	if biltyIDStr == "" {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Missing bilty ID",
		})
		return
	}

	biltyID, err := strconv.ParseInt(biltyIDStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid bilty ID",
		})
		return
	}

	if err := h.Repo.DeleteBilty(biltyID); err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to delete bilty: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Bilty deleted successfully",
	})
}
