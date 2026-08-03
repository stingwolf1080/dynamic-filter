package db

import (
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
)

type Connection interface {
	Connect(opts Options) error
	Create(model any, table string) error
	CreateMany(model []any, table string) error
	Read(filters filter.FilterOptions, table string, result any) error
	ReadMany(filters filter.FilterOptions, table string, results any) error
	Update(filters filter.FilterOptions, data any, table string) error
	UpdateMany(filters filter.FilterOptions, data any, table string) error
	Delete(filters filter.FilterOptions, table string) error
	DeleteMany(filters filter.FilterOptions, table string) error
	Count(filters filter.FilterOptions, table string) error
	CheckNotFound() error
	CreateTextIndex(table string, fields []string) error
	Close() error
}
