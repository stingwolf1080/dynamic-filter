package mongox

import (
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
	"go.mongodb.org/mongo-driver/bson"
)

func convertFilter(count bool, filters *filter.FilterOptions) (filter bson.M) {
	filter = make(bson.M)
	if filters.Search != "" {
		filter["$text"] = bson.M{"$search": `"` + filters.Search + `"`}
	}
	if !count {
		if filters.LastId() {
			filter["_id"] = bson.M{"$gt": filters.NextId()}
		}
	}

	for _, v := range filters.Filter {
		if len(v.Value) <= 0 {
			continue
		}
		if v.OrOperator {
			var orFunc bson.A
			for _, f_v := range v.Value {
				orFunc = append(orFunc, bson.M{filters.CheckAlias(f_v.Field): bson.M{"$" + f_v.Operator: f_v.ConvertValue(filters.Timezone())}})
			}
			filter["$or"] = orFunc
			continue
		}
		if len(v.Value) > 1 {
			var andFunc bson.A
			for _, f_v := range v.Value {
				andFunc = append(andFunc, bson.M{filters.CheckAlias(f_v.Field): bson.M{"$" + f_v.Operator: f_v.ConvertValue(filters.Timezone())}})
			}
			filter["$and"] = andFunc
			continue
		}
		value := v.Value[0]
		filter[filters.CheckAlias(value.Field)] = bson.M{"$" + value.Operator: value.ConvertValue(filters.Timezone())}
	}
	return filter
}
