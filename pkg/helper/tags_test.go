package helper

import (
	"testing"
)

type TestModel struct {
	ID       string `json:"id" bson:"_id,omitempty"`
	Name     string `json:"name" bson:"name,text"`
	Age      int    `json:"age" bson:"age"`
	Internal string `bson:"internal"` // no json
	Ignored  string `json:"-" bson:"-"`
}

func TestGetModelTags(t *testing.T) {
	tags := GetModelTags(TestModel{})

	// Check JSON tags
	if tags.JsonTags["ID"] != "id" {
		t.Errorf("Expected JsonTag 'id' for 'ID', got '%s'", tags.JsonTags["ID"])
	}
	if tags.JsonTags["Name"] != "name" {
		t.Errorf("Expected JsonTag 'name' for 'Name', got '%s'", tags.JsonTags["Name"])
	}

	// Check BSON tags
	if tags.BsonTags["ID"] != "_id" {
		t.Errorf("Expected BsonTag '_id' for 'ID', got '%s'", tags.BsonTags["ID"])
	}
	if tags.BsonTags["Name"] != "name" {
		t.Errorf("Expected BsonTag 'name' for 'Name', got '%s'", tags.BsonTags["Name"])
	}
	if tags.BsonTags["Internal"] != "internal" {
		t.Errorf("Expected BsonTag 'internal' for 'Internal', got '%s'", tags.BsonTags["Internal"])
	}

	// Check ignored fields
	if _, ok := tags.JsonTags["Ignored"]; ok {
		t.Errorf("Expected 'Ignored' field to not have a JsonTag")
	}

	// Check Text tags
	foundText := false
	for _, textTag := range tags.TextTags {
		if textTag == "name" {
			foundText = true
			break
		}
	}
	if !foundText {
		t.Errorf("Expected 'name' to be in TextTags")
	}
}
