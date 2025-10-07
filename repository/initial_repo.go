package repository

import "hariomtransport/models"

type InitialRepository interface {
	SaveInitial(initial *models.InitialSetup) error
	GetInitial() (*models.InitialSetup, error)
}
