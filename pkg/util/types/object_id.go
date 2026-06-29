package types

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ObjectID string

func NewObjectID() ObjectID {
	return ObjectID(primitive.NewObjectID().Hex())
}

func (id ObjectID) IsZero() bool {
	return id == "" || id == "000000000000000000000000"
}

func (id ObjectID) ObjectID() (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(string(id))
}

func (id ObjectID) String() string {
	return string(id)
}

func (id ObjectID) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if id.IsZero() {
		return bson.MarshalValue(primitive.NilObjectID) // hoặc bsontype.Null
	}
	objID, err := primitive.ObjectIDFromHex(string(id))
	if err != nil {
		return bsontype.Null, nil, err
	}
	return bson.MarshalValue(objID)
}

func (id *ObjectID) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	if t == bsontype.Null {
		*id = "" // hoặc "000000000000000000000000" nếu muốn zero
		return nil
	}
	// log.Println(t)
	var objID primitive.ObjectID
	if err := bson.UnmarshalValue(t, data, &objID); err != nil {
		return err
	}
	*id = ObjectID(objID.Hex())
	return nil
}

func IsObjectID(input any) (primitive.ObjectID, error) {
	switch v := input.(type) {
	case primitive.ObjectID:
		return v, nil
	case string:
		return primitive.ObjectIDFromHex(v)
	default:
		return primitive.NilObjectID, fmt.Errorf("invalid type: %T", input)
	}
}
