package repository

import "hariomtransport/models"

type BiltyRepository interface {
	CreateBiltyWithParties(bilty *models.Bilty) error
	GetBilty(filters map[string]interface{}, single bool) ([]*models.Bilty, error)
}
