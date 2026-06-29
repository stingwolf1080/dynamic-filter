package basemodel

import (
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseModel struct {
	Id        types.ID  `bson:"_id" json:"id"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
	Returned  bool      `bson:"-" json:"-"`
	NoneID    bool      `bson:"-" json:"-"`
}

func (b *BaseModel) NewID(id ...types.ID) {
	if len(id) > 0 {
		b.Id = id[0]
	} else {
		b.Id = types.NewID(primitive.NewObjectID())
	}
}
func (b *BaseModel) NewModel(id ...types.ID) {
	b.NewID(id...)
	b.CreatedAt = time.Now()
	b.UpdatedAt = time.Now()
}
func (b BaseModel) HasReturned() bool                 { return b.Returned }
func (b BaseModel) GetID() (bool, types.ID) { return b.NoneID, b.Id }
func (b BaseModel) GetUpdatedAt() time.Time           { return b.UpdatedAt }
func (b *BaseModel) SetUpdatedAt()                    { b.UpdatedAt = time.Now() }

func (b *BaseModel) SetID(id types.ID) (err error) {
	switch v := id.Val.(type) {
	case string:
		if oid, err := primitive.ObjectIDFromHex(v); err == nil {
			b.Id = types.NewID(oid)
		} else if parsedUUID, err := uuid.Parse(v); err == nil {
			b.Id = types.NewID(parsedUUID)
		} else if intId, err := strconv.Atoi(v); err == nil {
			b.Id = types.NewID(intId)
		} else {
			b.Id = types.NewID(v)
		}
	default:
		b.Id = types.NewID(v)
	}
	return nil
}

type BaseEntity interface {
	GetID() (bool, types.ID)
	SetID(id types.ID) error
	NewModel(id ...types.ID)
	SetUpdatedAt()
	GetPrefix() string
}
