package repository

import (
	"context"

	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type Repository[T any] interface {
	// default action modify data
	Create(data T) (message types.Message)
	Update(data T) (message types.Message)
	GenericDeleter

	GetByFilter(filter string) (message types.Message)
	CheckFilter(filter string) (message types.Message)
	ListPage(filter string) (message types.Message)
	// Aggregate(filter []bson.M) (message types.Message)
	// handle default data
	SetRespone(data any)
	SetIsReturn()
	SetTimezone(zone string)
	// SetID(id types.ID)
	// register func hook
	RegisterHandle(name string, fn func(ctx context.Context, data any, prefix string) types.Message)
	RegisterModel()
	// call hook function model repository
	CallFunc(name string, data any) (message types.Message)
	// new func update many
	UpdateMany(filter string, data map[string]any) (message types.Message)
	CreateMany(data []T) (message types.Message)
	DeleteMany(filter string, delete_message types.DeletePost) (message types.Message)
	CreateTextIndex(fields ...string) error
}

type GenericDeleter interface {
	Delete(delete_message types.DeletePost) (message types.Message)
	SetPrefix(prefix string)
}
