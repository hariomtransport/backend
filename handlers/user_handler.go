package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/hariomtransport/models"
	"github.com/hariomtransport/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Repo repository.UserRepository
}

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var user models.AppUser
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Email == "" || user.Role == "" {
		http.Error(w, "Name, email, and role are required", http.StatusBadRequest)
		return
	}

	if err := h.Repo.CreateUser(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.Repo.GetUserByEmail(creds.Email)
	if err != nil || user == nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	user.Password = "" // hide hash
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
