package mongox

import (
	"strconv"

	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func convertOptions(filters *filter.FilterOptions) *options.FindOptions {
	opts := &options.FindOptions{}
	if filters.Pagination() {
		if !filters.LastId() {
			opts.SetSkip(filters.Skip())
		}
		opts.SetLimit(filters.Limit())
	}
	if len(filters.Sort) > 0 {
		var sort bson.D
		for _, v := range filters.Sort {
			sort = append(sort, bson.E{
				Key:   filters.CheckAlias(v.Field),
				Value: v.Index,
			})
		}
		opts.SetSort(sort)
	}
	if len(filters.Fields) > 0 {
		var projection bson.D
		for _, v := range filters.Fields {
			key := v.FieldName
			var val any
			val = 1
			if v.IsValue {
				if i, err := strconv.Atoi(v.Value); err == nil {
					val = bson.M{"$literal": i}
				} else if f, err := strconv.ParseFloat(v.Value, 64); err == nil {
					val = bson.M{"$literal": f}
				} else {
					val = v.Value
				}
			}
			projection = append(projection, bson.E{
				Key:   key,
				Value: val,
			})
		}
		opts.SetProjection(projection)
	}
	return opts
}
