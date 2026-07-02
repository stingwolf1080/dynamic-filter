package repository

import (
	"context"
	"fmt"
	"math"

	hooks "github.com/stingwolf1080/dynamic-filter/hook"
	"github.com/stingwolf1080/dynamic-filter/pkg/db"
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"github.com/stingwolf1080/dynamic-filter/pkg/mongox"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type repository[T any] struct {
	collectionName string
	modelName      string
	prefix         string
	restrict       map[string]bool
	alias          map[string]string
	isReturn       bool
	timezone       string
	response       any
	dbConn         db.Connection
	// queryCache     *cache.Cache
}

func NewRepository[T any](dbConn db.Connection, restrict map[string]bool, alias map[string]string) Repository[T] {
	modelName := helper.GetNameModel[T]()
	r := &repository[T]{
		collectionName: helper.GetModelTableGeneric[T](),
		modelName:      modelName,
		restrict:       restrict,
		alias:          alias,
		dbConn:         dbConn,
		// queryCache:     cache.NewCache(),
	}
	RegisterDeleteFunc(modelName, r)
	return r
}

func (r *repository[T]) CheckFilter(filterStr string) (message types.Message) {
	opts_filter := &filter.FilterOptions{RestrictField: r.restrict}
	err := filter.ParseBracketParams(filterStr, opts_filter)
	if err != nil {
		message.Status = "error"
		message.Code = 400
		message.Message = err.Error()
		message.ErrorCode = types.ErrSystemParseFilter
		return
	}
	message.Status = "success"
	message.Code = 200
	return
}

func (r *repository[T]) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *repository[T]) NewEntity() T {
	var entity T
	return entity
}

func (r *repository[T]) SetRespone(data any) {
	r.response = data
}

func (r *repository[T]) SetIsReturn() {
	r.isReturn = true
}

func (r *repository[T]) GetCollection() string {
	if r.prefix == "" {
		return r.collectionName
	}
	return r.prefix + "_" + r.collectionName
}

func (r *repository[T]) SetTimezone(zone string) {
	r.timezone = zone
}

func (r *repository[T]) RegisterHandle(name string, fn func(ctx context.Context, data any, prefix string) types.Message) {
	registerHandle(r.modelName, name, fn)
}

func (r *repository[T]) RegisterModel() {
	mongox.Register(r.NewEntity(), true)
}

func (r *repository[T]) CallFunc(name string, data any) (message types.Message) {
	return callFunc(context.Background(), r.modelName, r.prefix, name, data)
}

func (r *repository[T]) Create(data T) (message types.Message) {
	if hook, ok := any(&data).(hooks.BeforeInsertHook); ok {
		if err := hook.BeforeInsert(); err != nil {
			return types.Message{
				Status:     "error",
				Code:       400,
				Message:    "Data insert error, please contact adminstrator",
				MessageErr: err,
				ErrorCode:  types.ErrSystemDatabase,
			}
		}
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeInsert, &data); err.HasError() {
		return err
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeSave, &data); err.HasError() {
		return err
	}
	if errDB := r.dbConn.Create(&data, r.GetCollection()); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database create failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}
	// r.queryCache.Clear()
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterInsert, &data); err.HasError() {
		return err
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterSave, &data); err.HasError() {
		return err
	}
	if r.isReturn {
		return types.Message{
			Status:  "success",
			Code:    201,
			Message: "Create successfuly!",
			Data:    data,
		}
	}
	return types.Message{
		Status:  "success",
		Code:    201,
		Message: "Create successfuly!",
	}
}

func (r *repository[T]) CreateMany(data []T) (message types.Message) {
	var docs []any
	var ids []types.ID
	for k, _data := range data {
		if hook, ok := any(&_data).(hooks.BeforeInsertHook); ok {
			if err := hook.BeforeInsert(); err != nil {
				return types.Message{
					Status:     "error",
					Code:       400,
					Message:    "Data insert error, please contact adminstrator",
					MessageErr: err,
					ErrorCode:  types.ErrSystemDatabase,
				}
			}
		}

		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeInsert, &_data); err.HasError() {
			return err
		}

		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeSave, &_data); err.HasError() {
			return err
		}
		data[k] = _data
		docs = append(docs, _data)
		if hook, ok := any(&_data).(hooks.HasID); ok {
			if none_id, id := hook.GetID(); !none_id {
				ids = append(ids, id)
			}
		}
	}
	if errDB := r.dbConn.CreateMany(docs, r.collectionName); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database create many failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}
	// r.queryCache.Clear()
	for k, _data := range data {
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterInsert, &_data); err.HasError() {
			return err
		}
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterSave, &_data); err.HasError() {
			return err
		}
		if hook, ok := any(&_data).(hooks.AfterInsertHook); ok {
			if err := hook.AfterInsert(ids); err != nil {
				return types.Message{
					Status:     "error",
					Code:       400,
					Message:    "Data insert error, please contact adminstrator",
					MessageErr: err,
					ErrorCode:  types.ErrSystemDatabase,
				}
			}
		}
		data[k] = _data

	}
	if r.isReturn {
		return types.Message{
			Status:  "success",
			Code:    201,
			Message: "Create successfuly!",
			Data:    data,
		}
	}
	return types.Message{
		Status:  "success",
		Code:    201,
		Message: "Create successfuly!",
	}
}

func (r *repository[T]) Update(data T) (message types.Message) {
	var r_id types.ID
	var none_id bool
	if hook, ok := any(&data).(hooks.HasID); ok {
		if none_id, r_id = hook.GetID(); !none_id {
			if r_id.IsZero() {
				return types.Message{
					Status:    "error",
					Code:      400,
					Message:   "Data does not exists",
					ErrorCode: types.ErrNotFoundData,
				}
			}
		}
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeUpdate, &data); err.HasError() {
		return err
	}

	if hook, ok := any(&data).(hooks.BeforeUpdateHook); ok {
		if err := hook.BeforeUpdate(); err != nil {
			return types.Message{
				Status:     "error",
				Code:       400,
				Message:    "Data update error, please contact adminstrator",
				MessageErr: err,
				ErrorCode:  types.ErrSystemDatabase,
			}
		}
		if none_id {
			if hook, ok := any(&data).(hooks.HasID); ok {
				_, r_id = hook.GetID()
				if r_id.IsZero() {
					return types.Message{
						Status:    "error",
						Code:      400,
						Message:   "Data does not exists",
						ErrorCode: types.ErrNotFoundData,
					}
				}
			}
		}
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeSave, &data); err.HasError() {
		return err
	}
	opts_filter := filter.FilterOptions{}
	if !none_id && !r_id.IsZero() {
		opts_filter.Filter = map[string]filter.Filter{
			"id": {
				Value: []filter.OperatorFilter{
					{Field: "id", Operator: "eq", Value: fmt.Sprint(r_id.Val)},
				},
			},
		}
	}
	if errDB := r.dbConn.Update(opts_filter, &data, r.collectionName); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database update failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}
	// r.queryCache.Clear()
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterUpdate, &data); err.HasError() {
		return err
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterSave, &data); err.HasError() {
		return err
	}
	if r.isReturn {
		return types.Message{
			Status:  "success",
			Code:    200,
			Message: "Update successfuly!",
			Data:    data,
		}
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Update successfuly!",
	}
}

func (r *repository[T]) Delete(delete_message types.DeletePost) (message types.Message) {
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &delete_message); err.HasError() {
		return err
	}

	if delete_message.AfterApprove {
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterApprove, &delete_message); err.HasError() {
			return err
		}
	}
	opts_filter := filter.FilterOptions{
		Filter: map[string]filter.Filter{
			"id": {
				Value: []filter.OperatorFilter{
					{Field: "id", Operator: "eq", Value: string(delete_message.Id)},
				},
			},
		},
	}
	if errDB := r.dbConn.Delete(opts_filter, r.collectionName); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database delete failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}
	// r.queryCache.Clear()
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterDelete, &delete_message); err.HasError() {
		return err
	}
	if r.isReturn {
		return types.Message{
			Status:  "success",
			Code:    200,
			Message: "Delete successfuly!",
		}
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Delete successfuly!",
	}
}

func (r *repository[T]) DeleteMany(filter_str string, delete_message types.DeletePost) (message types.Message) {
	opts_filter := &filter.FilterOptions{AliasField: r.alias}
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}
	var err error
	err = filter.ParseBracketParams(filter_str, opts_filter)
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Unable to parse: an object hierarchy has been provided",
			MessageErr: err,
			ErrorCode:  types.ErrSystemParseFilter,
		}
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &delete_message); err.HasError() {
		return err
	}
	if _, ok := any(&delete_message).(hooks.HasDeletePost); ok {
		if delete_message.AfterApprove {
			if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterApprove, &delete_message); err.HasError() {
				return err
			}
		}
	}

	if errDB := r.dbConn.DeleteMany(*opts_filter, r.collectionName); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database delete many failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}
	// r.queryCache.Clear()
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Data delete error, please contact adminstrator",
			MessageErr: err,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterDelete, &delete_message); err.HasError() {
		return err
	}
	if r.isReturn {
		return types.Message{
			Status:  "success",
			Code:    200,
			Message: "Delete successfuly!",
			// Data:    resulf,
		}
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Delete successfuly!",
	}
}

func (r *repository[T]) GetByFilter(filterStr string) (message types.Message) {
	// if val, ok := r.queryCache.Get("get:" + filterStr); ok {
	// 	return val.(types.Message)
	// }

	opts_filter := &filter.FilterOptions{AliasField: r.alias}
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}

	err := filter.ParseBracketParams(filterStr, opts_filter)
	if err != nil {
		message.Status = "error"
		message.Code = 400
		message.Message = "Unable to parse: an object hierarchy has been provided"
		message.MessageErr = err
		message.ErrorCode = types.ErrSystemParseFilter
		return message
	}
	var result any = r.response
	if result == nil {
		var defaultResult []T
		result = &defaultResult
		r.response = &defaultResult
	}
	if errDB := r.dbConn.ReadMany(*opts_filter, r.collectionName, result); errDB != nil {
		message.Status = "error"
		message.Code = 500
		message.Message = "Database read failed"
		message.MessageErr = errDB
		message.ErrorCode = types.ErrSystemDatabase
		return message
	}

	msg := types.Message{
		Status: "success",
		Code:   200,
		Data:   r.response,
	}
	// r.queryCache.Set("get:"+filterStr, msg)
	return msg
}

func (r *repository[T]) ListPage(filterStr string) (message types.Message) {
	// if val, ok := r.queryCache.Get("list:" + filterStr); ok {
	// 	return val.(types.Message)
	// }

	opts_filter := &filter.FilterOptions{AliasField: r.alias}
	opts_filter.SetPagination()
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}
	err := filter.ParseBracketParams(filterStr, opts_filter)
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Unable to parse: an object hierarchy has been provided",
			MessageErr: err,
			ErrorCode:  types.ErrSystemParseFilter,
		}
	}

	var result any = r.response
	if result == nil {
		var defaultResult []T
		result = &defaultResult
		r.response = &defaultResult
	}
	if errDB := r.dbConn.ReadMany(*opts_filter, r.collectionName, result); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database read failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}

	msg := types.Message{
		Status:    "success",
		Code:      200,
		Data:      r.response,
		TotalRow:  100,
		TotalPage: int64(math.Ceil(float64(100) / float64(opts_filter.Limit()))),
	}
	// r.queryCache.Set("list:"+filterStr, msg)
	return msg
}

func (r *repository[T]) UpdateMany(filterStr string, data map[string]any) (message types.Message) {
	if len(data) <= 0 {
		return types.Message{
			Status:  "error",
			Code:    400,
			Message: "Data update error, please contact adminstrator",
		}
	}

	opts_filter := filter.FilterOptions{AliasField: r.alias}
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}

	err := filter.ParseBracketParams(filterStr, &opts_filter)
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Unable to parse filter",
			MessageErr: err,
			ErrorCode:  types.ErrSystemParseFilter,
		}
	}

	if errDB := r.dbConn.UpdateMany(opts_filter, data, r.collectionName); errDB != nil {
		return types.Message{
			Status:     "error",
			Code:       500,
			Message:    "Database update many failed",
			MessageErr: errDB,
			ErrorCode:  types.ErrSystemDatabase,
		}
	}

	// r.queryCache.Clear()

	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Update successfuly!",
	}
}
