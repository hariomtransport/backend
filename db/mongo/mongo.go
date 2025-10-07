package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client *mongo.Client
	Ctx    context.Context
	Cancel context.CancelFunc
	URL    string
}

func NewMongoDB(url string) *MongoDB {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &MongoDB{
		Ctx:    ctx,
		Cancel: cancel,
		URL:    url,
	}
}

func (m *MongoDB) Connect() error {
	client, err := mongo.Connect(m.Ctx, options.Client().ApplyURI(m.URL))
	if err != nil {
		return err
	}
	m.Client = client
	return m.Client.Ping(m.Ctx, nil)
}

func (m *MongoDB) Disconnect() error {
	m.Cancel()
	return m.Client.Disconnect(m.Ctx)
}

func (m *MongoDB) GetContext() context.Context {
	return m.Ctx
}
