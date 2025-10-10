package repository

import (
	"github.com/hariomtransport/models"
)

type InitialRepository interface {
	SaveInitial(initial *models.InitialSetup) error
	GetInitial() (*models.InitialSetup, error)
}
