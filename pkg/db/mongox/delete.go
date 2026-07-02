package mongox

import (
	"context"
	"time"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

func (conn *Conn) Delete(filters filter.FilterOptions, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	f := convertFilter(false, &filters)
	_, err := dataCollection.DeleteOne(ctx, f)
	return err
}

func (conn *Conn) DeleteMany(filters filter.FilterOptions, table string) error {
	dataCollection := conn.Database.Collection(table)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	f := convertFilter(false, &filters)
	_, err := dataCollection.DeleteMany(ctx, f)
	return err
}
