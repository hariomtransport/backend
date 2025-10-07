package models

import "time"

type AppUser struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`
	Password  string    `json:"password" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
