package models

import "time"

// Snapshot address stored per bilty to preserve historical data.
type BiltyAddress struct {
	ID          int64     `json:"id" db:"id"`
	CompanyID   *int64    `json:"company_id,omitempty" db:"company_id"`
	AddressLine string    `json:"address_line" db:"address_line"`
	City        string    `json:"city" db:"city"`
	State       string    `json:"state" db:"state"`
	Pincode     string    `json:"pincode" db:"pincode"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
