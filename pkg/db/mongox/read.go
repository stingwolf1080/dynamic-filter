package mongox

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (conn *Conn) Read(filter bson.M, table string, opts *options.FindOneOptions, result interface{}) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	err = dataCollection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		return err
	}

	return nil
}

func (conn *Conn) ReadMany(filter bson.M, table string, opts *options.FindOptions, results interface{}) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	cursor, err := dataCollection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, results); err != nil {
		return err
	}
	return nil
}

func (conn *Conn) Count(filter bson.M, table string, opts *options.CountOptions) (count int64, err error) {
	dataCollection := conn.Database.Collection(table)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	count, err = dataCollection.CountDocuments(ctx, filter, opts)
	if err != nil {
		return count, err
	}
	return count, nil
}

func (conn *Conn) AggregateOne(filter []bson.M, collection string, opts *options.AggregateOptions, results interface{}) (err error) {
	dataCollection := conn.Database.Collection(collection)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	cursor, err := dataCollection.Aggregate(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if cursor.Next(ctx) {
		if err := cursor.Decode(results); err != nil {
			return err
		}
	} else {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (conn *Conn) Aggregate(filter []bson.M, collection string, opts *options.AggregateOptions, results interface{}) (err error) {
	dataCollection := conn.Database.Collection(collection)
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
	cursor, err := dataCollection.Aggregate(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, results); err != nil {
		return err
	}
	return nil
}
