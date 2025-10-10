package models

import "time"

type Bilty struct {
	ID                 int64      `json:"id" db:"id"`
	BiltyNo            int64      `json:"bilty_no" db:"bilty_no"`
	ConsignorCompanyID *int64     `json:"consignor_company_id,omitempty" db:"consignor_company_id"`
	ConsigneeCompanyID *int64     `json:"consignee_company_id,omitempty" db:"consignee_company_id"`
	ConsignorAddressID *int64     `json:"consignor_address_id,omitempty" db:"consignor_address_id"`
	ConsigneeAddressID *int64     `json:"consignee_address_id,omitempty" db:"consignee_address_id"`
	FromLocation       string     `json:"from_location" db:"from_location"`
	ToLocation         string     `json:"to_location" db:"to_location"`
	Date               time.Time  `json:"date" db:"date"`
	ToPay              float64    `json:"to_pay" db:"to_pay"`
	GSTIN              *string    `json:"gstin,omitempty" db:"gstin"`
	InvNo              *string    `json:"inv_no,omitempty" db:"inv_no"`
	PVTMarks           *string    `json:"pvt_marks,omitempty" db:"pvt_marks"`
	PermitNo           *string    `json:"permit_no,omitempty" db:"permit_no"`
	ValueRupees        *float64   `json:"value_rupees,omitempty" db:"value_rupees"`
	Remarks            *string    `json:"remarks,omitempty" db:"remarks"`
	Hamali             *float64   `json:"hamali,omitempty" db:"hamali"`
	DDCharges          *float64   `json:"dd_charges,omitempty" db:"dd_charges"`
	OtherCharges       *float64   `json:"other_charges,omitempty" db:"other_charges"`
	FOV                *float64   `json:"fov,omitempty" db:"fov"`
	Statistical        *string    `json:"statistical,omitempty" db:"statistical"`
	CreatedBy          int64      `json:"created_by" db:"created_by"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at" db:"updated_at"`
	PdfCreatedAt       *time.Time `json:"pdf_created_at" db:"pdf_created_at"`
	PdfPath            *string    `json:"pdf_path,omitempty" db:"pdf_path"`
	Status             string     `json:"status" db:"status"` // draft | complete

	// Nested objects for responses (denormalized)
	ConsignorCompany     *Company      `json:"consignor_company,omitempty"`
	ConsigneeCompany     *Company      `json:"consignee_company,omitempty"`
	ConsignorAddressSnap *BiltyAddress `json:"consignor_address_snapshot,omitempty"`
	ConsigneeAddressSnap *BiltyAddress `json:"consignee_address_snapshot,omitempty"`
	CreatedByUser        *AppUser      `json:"created_by_user,omitempty"`
	Goods                []Goods       `json:"goods,omitempty"`
}
