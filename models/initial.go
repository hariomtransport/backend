package models

import "time"

type MobileEntry struct {
	Number string `json:"number" bson:"number" db:"number"`
	Label  string `json:"label" bson:"label" db:"label"`
}

type InitialSetup struct {
	ID          int64         `json:"id" bson:"_id,omitempty" db:"id"`
	CompanyName string        `json:"company_name" bson:"name" db:"name"`
	Address     string        `json:"address" bson:"address" db:"address"`
	City        string        `json:"city" bson:"city" db:"city"`
	State       string        `json:"state" bson:"state" db:"state"`
	Pincode     string        `json:"pincode" bson:"pincode" db:"pincode"`
	GSTIN       string        `json:"gstin" bson:"gstin" db:"gstin"`
	Footnote    string        `json:"footnote" bson:"footnote" db:"footnote"`
	Mobile      []MobileEntry `json:"mobile" bson:"mobile" db:"mobile"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at" db:"created_at"`
}
