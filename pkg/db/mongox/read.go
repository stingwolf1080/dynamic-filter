package mongox

import (
	"context"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
	"go.mongodb.org/mongo-driver/mongo"
)

func (conn *Conn) Read(filters filter.FilterOptions, table string, result any) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	f := convertFilter(false, &filters)
	err = dataCollection.FindOne(ctx, f).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return db.ErrNotFound
		}
		return err
	}
	return nil
}

func (conn *Conn) ReadMany(filters filter.FilterOptions, table string, results any) error {
	var err error
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	f := convertFilter(false, &filters)
	opts := convertOptions(&filters)

	cursor, err := dataCollection.Find(ctx, f, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	if err = cursor.All(ctx, results); err != nil {
		return err
	}
	return nil
}

func (conn *Conn) Count(filters filter.FilterOptions, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	f := convertFilter(true, &filters)
	// We execute count but we don't return it because interface doesn't return count.
	// We'll just verify syntax.
	_, err := dataCollection.CountDocuments(ctx, f, nil)
	return err
}
