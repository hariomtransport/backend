package repository

import (
	"time"

	"github.com/hariomtransport/backend/models"
)

type BiltyRepository interface {
	CreateBiltyWithParties(bilty *models.Bilty) error
	GetBilty(filters map[string]interface{}, single bool) ([]*models.Bilty, error)
	UpdatePDFInfo(biltyID int64, pdfPath string, t time.Time) error
	DeleteBilty(biltyID int64) error
	GetBiltyByID(biltyID int64) (*models.Bilty, error)
}
