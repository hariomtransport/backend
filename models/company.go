package models

import "time"

type Company struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	GSTIN     *string   `json:"gstin,omitempty" db:"gstin"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
