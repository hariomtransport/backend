package repository

import (
	"github.com/hariomtransport/backend/models"
)

type InitialRepository interface {
	SaveInitial(initial *models.InitialSetup) error
	GetInitial() (*models.InitialSetup, error)
}
