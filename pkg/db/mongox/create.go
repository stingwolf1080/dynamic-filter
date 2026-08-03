package mongox

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (conn *Conn) CreateTextIndex(table string, fields []string) error {
	if len(fields) == 0 {
		return nil
	}

	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var keysD bson.D
	for _, field := range fields {
		keysD = append(keysD, bson.E{Key: field, Value: "text"})
	}

	indexModel := mongo.IndexModel{
		Keys: keysD,
	}

	_, err := dataCollection.Indexes().CreateOne(ctx, indexModel)
	return err
}
