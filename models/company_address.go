package models

import "time"

type CompanyAddress struct {
	ID          int64     `json:"id" db:"id"`
	CompanyID   int64     `json:"company_id" db:"company_id"`
	AddressLine string    `json:"address_line" db:"address_line"`
	City        string    `json:"city" db:"city"`
	State       string    `json:"state" db:"state"`
	Pincode     string    `json:"pincode" db:"pincode"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
