package repository

import (
	"github.com/hariomtransport/backend/models"
)

// PDFRepository provides methods to fetch data for PDF generation
type PDFRepository struct {
	BiltyRepo   BiltyRepository
	InitialRepo InitialRepository
}

// NewPDFRepository initializes a PDF repository
func NewPDFRepository(biltyRepo BiltyRepository, initialRepo InitialRepository) *PDFRepository {
	return &PDFRepository{
		BiltyRepo:   biltyRepo,
		InitialRepo: initialRepo,
	}
}

// GetBiltyForPDF fetches a single bilty by ID for PDF
func (r *PDFRepository) GetBiltyForPDF(id int64) (*models.Bilty, error) {
	bilties, err := r.BiltyRepo.GetBilty(map[string]interface{}{"id": id}, true)
	if err != nil {
		return nil, err
	}
	if len(bilties) == 0 {
		return nil, nil
	}
	return bilties[0], nil
}

// GetInitialForPDF fetches the latest initial setup / company info
func (r *PDFRepository) GetInitialForPDF() (*models.InitialSetup, error) {
	return r.InitialRepo.GetInitial()
}
