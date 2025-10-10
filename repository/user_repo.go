package repository

import "github.com/hariomtransport/models"

// UserRepository defines the interface for user operations
type UserRepository interface {
	CreateUser(user *models.AppUser) error
	GetUserByEmail(email string) (*models.AppUser, error)
}
