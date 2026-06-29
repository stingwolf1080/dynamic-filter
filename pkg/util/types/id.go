package types

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDType string

const (
	IDTypeUnknown  IDType = ""
	IDTypeUUID     IDType = "uuid"
	IDTypeInt      IDType = "int"
	IDTypeObjectID IDType = "objectid"
	IDTypeString   IDType = "string"
)

// ID is a flexible identifier type that can hold uuid, int, string or primitive.ObjectID
type ID struct {
	Val  any
	Type IDType
}

func NewID(val any) ID {
	switch v := val.(type) {
	case primitive.ObjectID:
		return ID{Val: v, Type: IDTypeObjectID}
	case uuid.UUID:
		return ID{Val: v, Type: IDTypeUUID}
	case int, int32, int64:
		return ID{Val: v, Type: IDTypeInt}
	case string:
		return ID{Val: v, Type: IDTypeString}
	default:
		return ID{Val: nil, Type: IDTypeUnknown}
	}
}

func (id ID) IsZero() bool {
	switch v := id.Val.(type) {
	case primitive.ObjectID:
		return v.IsZero()
	case uuid.UUID:
		return v == uuid.Nil
	case string:
		return v == ""
	case int, int32, int64:
		return v == 0
	default:
		return false
	}
}

func (id ID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if id.Val == nil {
		return bsontype.Null, nil, nil
	}
	return bson.MarshalValue(id.Val)
}

func (id *ID) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.Null {
		id.Val = nil
		id.Type = IDTypeUnknown
		return nil
	}
	var v any
	err := bson.UnmarshalValue(t, data, &v)
	if err != nil {
		return err
	}
	*id = NewID(v)
	return nil
}

func (id ID) Value() (driver.Value, error) {
	return id.Val, nil
}

func (id *ID) Scan(value any) error {
	*id = NewID(value)
	return nil
}

func (id ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.Val)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	// Handle float64 from JSON numbers
	if f, ok := v.(float64); ok && f == float64(int(f)) {
		*id = NewID(int(f))
	} else {
		*id = NewID(v)
	}
	return nil
}
