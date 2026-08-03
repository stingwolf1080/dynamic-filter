package helper

import (
	"reflect"
	"strings"
	"sync"
)

type ModelTags struct {
	JsonTags  map[string]string
	BsonTags  map[string]string
	AliasTags map[string]string
	TextTags  []string
}

var modelTagsCache sync.Map

// GetModelTags parses and caches struct tags for a given model.
func GetModelTags(model any) ModelTags {
	modelName := GetModelName(model)
	if cached, ok := modelTagsCache.Load(modelName); ok {
		return cached.(ModelTags)
	}

	tags := ModelTags{
		JsonTags:  make(map[string]string),
		BsonTags:  make(map[string]string),
		AliasTags: make(map[string]string),
		TextTags:  make([]string, 0),
	}
	
	extractTags(reflect.TypeOf(model), &tags)

	modelTagsCache.Store(modelName, tags)
	return tags
}

func extractTags(t reflect.Type, tags *ModelTags) {
	if t == nil {
		return
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		if field.Anonymous {
			extractTags(field.Type, tags)
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			parts := strings.Split(jsonTag, ",")
			tags.JsonTags[field.Name] = parts[0]
		}

		bsonTag := field.Tag.Get("bson")
		if bsonTag != "" && bsonTag != "-" {
			parts := strings.Split(bsonTag, ",")
			tags.BsonTags[field.Name] = parts[0]
			for _, part := range parts {
				if part == "text" {
					tags.TextTags = append(tags.TextTags, parts[0])
				}
			}
		}

		aliasTag := field.Tag.Get("alias")
		if aliasTag != "" && aliasTag != "-" {
			parts := strings.Split(aliasTag, ",")
			tags.AliasTags[field.Name] = parts[0]
		}
	}
}
