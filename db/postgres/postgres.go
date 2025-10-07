package postgres

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	Conn   *sql.DB
	Ctx    context.Context
	Cancel context.CancelFunc
	URL    string
}

func NewPostgresDB(url string) *PostgresDB {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return &PostgresDB{
		Ctx:    ctx,
		Cancel: cancel,
		URL:    url,
	}
}

func (p *PostgresDB) Connect() error {
	conn, err := sql.Open("postgres", p.URL)
	if err != nil {
		return err
	}
	p.Conn = conn
	return p.Conn.Ping()
}

func (p *PostgresDB) Disconnect() error {
	p.Cancel()
	return p.Conn.Close()
}

func (p *PostgresDB) GetContext() context.Context {
	return p.Ctx
}
