package mongox

import (
	"reflect"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type textField struct {
	key    string
	weight int
}

func parseModel(model any) ([]mongo.IndexModel, error) {
	t := reflect.TypeOf(model)

	var normalIndexes []mongo.IndexModel
	var textFields []textField

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		bsonTag := f.Tag.Get("bson")
		if bsonTag == "" {
			continue
		}

		key := strings.Split(bsonTag, ",")[0]

		// TEXT INDEX
		if tag := f.Tag.Get("text"); tag != "" {
			weight := 1
			if tag != "" {
				if w, err := strconv.Atoi(tag); err == nil {
					weight = w
				}
			}

			textFields = append(textFields, textField{
				key:    key,
				weight: weight,
			})
			continue
		}

		// NORMAL INDEX
		switch f.Tag.Get("idx") {

		case "1":
			normalIndexes = append(normalIndexes, mongo.IndexModel{
				Keys: bson.D{{Key: key, Value: 1}},
			})

		case "-1":
			normalIndexes = append(normalIndexes, mongo.IndexModel{
				Keys: bson.D{{Key: key, Value: -1}},
			})

		case "unique":
			normalIndexes = append(normalIndexes, mongo.IndexModel{
				Keys:    bson.D{{Key: key, Value: 1}},
				Options: options.Index().SetUnique(true),
			})

		case "2dsphere":
			normalIndexes = append(normalIndexes, mongo.IndexModel{
				Keys: bson.D{{Key: key, Value: "2dsphere"}},
			})
		}
	}

	// BUILD TEXT INDEX (1 per collection)
	var indexes []mongo.IndexModel
	indexes = append(indexes, normalIndexes...)

	if len(textFields) > 0 {
		keys := bson.D{}
		weights := bson.D{}

		for _, f := range textFields {
			keys = append(keys, bson.E{
				Key:   f.key,
				Value: "text",
			})

			weights = append(weights, bson.E{
				Key:   f.key,
				Value: f.weight,
			})
		}

		indexes = append(indexes, mongo.IndexModel{
			Keys:    keys,
			Options: options.Index().SetWeights(weights),
		})
	}

	return indexes, nil
}
