package mongox

import (
	"context"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
	"go.mongodb.org/mongo-driver/bson"
)

func (conn *Conn) Update(filters filter.FilterOptions, data any, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	f := convertFilter(false, &filters)
	update := bson.M{"$set": data}
	_, err := dataCollection.UpdateOne(ctx, f, update)
	return err
}

func (conn *Conn) UpdateMany(filters filter.FilterOptions, data any, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	f := convertFilter(false, &filters)
	update := bson.M{"$set": data}
	_, err := dataCollection.UpdateMany(ctx, f, update)
	return err
}
