package db

import "github.com/stingwolf1080/dynamic-filter/pkg/filter"

type Wrapper struct {
	conn Connection
}

func (w *Wrapper) Connect(opts Options) error {
	return w.conn.Connect(opts)
}

func (w *Wrapper) Create(model any, table string) error {
	return w.conn.Create(model, table)
}

func (w *Wrapper) CreateMany(model []any, table string) error {
	return w.conn.CreateMany(model, table)
}

func (w *Wrapper) Read(filters filter.FilterOptions, table string, result any) error {
	return w.conn.Read(filters, table, result)
}

func (w *Wrapper) ReadMany(filters filter.FilterOptions, table string, results any) error {
	return w.conn.ReadMany(filters, table, results)
}

func (w *Wrapper) Update(filters filter.FilterOptions, data any, table string) error {
	return w.conn.Update(filters, data, table)
}

func (w *Wrapper) UpdateMany(filters filter.FilterOptions, data any, table string) error {
	return w.conn.UpdateMany(filters, data, table)
}

func (w *Wrapper) Delete(filters filter.FilterOptions, table string) error {
	return w.conn.Delete(filters, table)
}

func (w *Wrapper) DeleteMany(filters filter.FilterOptions, table string) error {
	return w.conn.DeleteMany(filters, table)
}

func (w *Wrapper) Count(filters filter.FilterOptions, table string) error {
	return w.conn.Count(filters, table)
}

func (w *Wrapper) CheckNotFound() error {
	return w.conn.CheckNotFound()
}

func (w *Wrapper) Close() error {
	return w.conn.Close()
}
