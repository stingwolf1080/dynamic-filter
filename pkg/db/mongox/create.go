package mongox

import (
	"context"
	"time"
)

func (conn *Conn) Create(data any, table string) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = dataCollection.InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Conn) CreateMany(data []any, table string) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	_, err = dataCollection.InsertMany(ctx, data)
	if err != nil {
		return err
	}
	return nil
}
