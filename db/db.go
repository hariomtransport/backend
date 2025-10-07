package db

import "context"

type DBType string

const (
	Postgres DBType = "postgres"
	Mongo    DBType = "mongo"
)

type DB interface {
	Connect() error
	Disconnect() error
	GetContext() context.Context
}
