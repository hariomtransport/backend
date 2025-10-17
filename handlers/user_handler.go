package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hariomtransport/backend/models"
	"github.com/hariomtransport/backend/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Repo repository.UserRepository
}

// Signup handler
func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ApiResponse{
			Success: false,
			Message: "Invalid request method",
		})
		return
	}

	var user models.AppUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}

	if user.Name == "" || user.Email == "" || user.Role == "" {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Name, email, and role are required",
		})
		return
	}

	if err := h.Repo.CreateUser(&user); err != nil {
		writeJSON(w, http.StatusInternalServerError, ApiResponse{
			Success: false,
			Message: "Failed to create user: " + err.Error(),
		})
		return
	}

	user.Password = "" // hide password

	writeJSON(w, http.StatusCreated, ApiResponse{
		Success: true,
		Message: "User signed up successfully",
		Data:    user,
	})
}

// Login handler
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ApiResponse{
			Success: false,
			Message: "Invalid request method",
		})
		return
	}

	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		writeJSON(w, http.StatusBadRequest, ApiResponse{
			Success: false,
			Message: "Invalid request payload: " + err.Error(),
		})
		return
	}

	user, err := h.Repo.GetUserByEmail(creds.Email)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, ApiResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, ApiResponse{
			Success: false,
			Message: "Invalid email or password",
		})
		return
	}

	user.Password = "" // hide password hash

	writeJSON(w, http.StatusOK, ApiResponse{
		Success: true,
		Message: "Login successful",
		Data:    user,
	})
}
