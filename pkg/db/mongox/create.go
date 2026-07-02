package mongox

import (
	"context"
	"time"
)

func (conn *Conn) Create(data any, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := dataCollection.InsertOne(ctx, data)
	return err
}

func (conn *Conn) CreateMany(data []any, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := dataCollection.InsertMany(ctx, data)
	return err
}
