package repository

import (
	"hariomtransport/models"
	"time"
)

type BiltyRepository interface {
	CreateBiltyWithParties(bilty *models.Bilty) error
	GetBilty(filters map[string]interface{}, single bool) ([]*models.Bilty, error)
	UpdatePDFCreatedAt(biltyID int64, t time.Time) error
	DeleteBilty(biltyID int64) error
}
