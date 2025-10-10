package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/hariomtransport/backend/models"

	"golang.org/x/crypto/bcrypt"
)

type PostgresUserRepo struct {
	DB *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) *PostgresUserRepo {
	return &PostgresUserRepo{DB: db}
}

// CreateUser creates a user after validating email uniqueness and hashing password
func (r *PostgresUserRepo) CreateUser(user *models.AppUser) error {
	// 1️⃣ Check if email already exists
	existingUser, err := r.GetUserByEmail(user.Email)
	if err != nil {
		return err
	}
	if existingUser != nil {
		return errors.New("email already exists")
	}

	// 2️⃣ Hash password
	if user.Password == "" {
		return errors.New("password cannot be empty")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	// 3️⃣ Set created_at if not set
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	// 4️⃣ Insert into DB
	_, err = r.DB.Exec(`
		INSERT INTO app_user (name, email, password, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, user.Name, user.Email, user.Password, user.Role, user.CreatedAt)

	return err
}

// GetUserByEmail fetches user by email
func (r *PostgresUserRepo) GetUserByEmail(email string) (*models.AppUser, error) {
	user := &models.AppUser{}
	err := r.DB.QueryRow(`
		SELECT id, name, email, password, role, created_at
		FROM app_user
		WHERE email=$1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
