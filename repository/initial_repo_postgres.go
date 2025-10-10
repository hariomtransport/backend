package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/hariomtransport/models"
)

type PostgresInitialRepo struct {
	DB *sql.DB
}

func NewPostgresInitialRepo(db *sql.DB) *PostgresInitialRepo {
	return &PostgresInitialRepo{DB: db}
}

// SaveInitial inserts or updates company initial details
func (r *PostgresInitialRepo) SaveInitial(initial *models.InitialSetup) error {
	if initial.CreatedAt.IsZero() {
		initial.CreatedAt = time.Now().UTC()
	}

	// Convert mobile slice to JSON manually
	mobileJSON, err := json.Marshal(initial.Mobile)
	if err != nil {
		return err
	}

	footnoteJSON, err := json.Marshal(initial.Footnote)
	if err != nil {
		return err
	}

	// If ID is passed → UPDATE, else INSERT
	if initial.ID > 0 {
		_, err = r.DB.Exec(`
			UPDATE initial_setup
			SET company_name=$1, gstin=$2, address=$3, city=$4, state=$5,
				pincode=$6, mobile=$7, footnote=$8, created_at=$9
			WHERE id=$10
		`, initial.CompanyName, initial.GSTIN, initial.Address, initial.City, initial.State,
			initial.Pincode, mobileJSON, footnoteJSON, initial.CreatedAt, initial.ID)
	} else {
		_, err = r.DB.Exec(`
			INSERT INTO initial_setup 
			(company_name, gstin, address, city, state, pincode, mobile, footnote, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, initial.CompanyName, initial.GSTIN, initial.Address, initial.City, initial.State,
			initial.Pincode, mobileJSON, footnoteJSON, initial.CreatedAt)
	}

	return err
}

// GetInitial fetches the latest initial setup
func (r *PostgresInitialRepo) GetInitial() (*models.InitialSetup, error) {
	initial := &models.InitialSetup{}
	var mobileJSON []byte
	var footnoteJSON []byte

	err := r.DB.QueryRow(`
		SELECT id, company_name, address, city, state, pincode, gstin, footnote, mobile, created_at
		FROM initial_setup
		ORDER BY id DESC LIMIT 1
	`).Scan(&initial.ID, &initial.CompanyName, &initial.Address, &initial.City, &initial.State,
		&initial.Pincode, &initial.GSTIN, &footnoteJSON, &mobileJSON, &initial.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Decode JSONB to Go slice
	if len(mobileJSON) > 0 {
		if err := json.Unmarshal(mobileJSON, &initial.Mobile); err != nil {
			return nil, err
		}
	}

	if len(footnoteJSON) > 0 {
		if err := json.Unmarshal(footnoteJSON, &initial.Footnote); err != nil {
			return nil, err
		}
	}

	return initial, nil
}
