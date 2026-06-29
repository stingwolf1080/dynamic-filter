package repository

import (
	"context"
	"math"

	"time"

	hooks "github.com/stingwolf1080/dynamic-filter/hook"
	"github.com/stingwolf1080/dynamic-filter/pkg/filter"
	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"github.com/stingwolf1080/dynamic-filter/pkg/mongox"
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type genericRepository[T any] struct {
	db             any
	collectionName string
	modelName      string
	prefix         string
	id             primitive.ObjectID
	restrict       map[string]bool
	alias          map[string]string
	isReturn       bool
	timezone       string
	response       any
}

func NewGenericRepository[T any](db any, collectionName string) Repository[T] {
	modelName := helper.GetNameModel[T]()
	return &genericRepository[T]{
		db:             db,
		collectionName: collectionName,
		modelName:      modelName,
		restrict:       make(map[string]bool),
		alias:          make(map[string]string),
	}
}

func (r *genericRepository[T]) CheckFilter(filterStr string) (message types.Message) {
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

func (r *genericRepository[T]) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *genericRepository[T]) NewEntity() T {
	var entity T
	return entity
}

func (r *genericRepository[T]) SetRespone(data any) {
	r.response = data
}

func (r *genericRepository[T]) SetIsReturn() {
	r.isReturn = true
}

func (r *genericRepository[T]) SetTimezone(zone string) {
	r.timezone = zone
}

func (r *genericRepository[T]) SetID(id types.ID) {
	if oid, ok := id.Val.(primitive.ObjectID); ok {
		r.id = oid
	} else if s, ok := id.Val.(string); ok {
		if oid, err := primitive.ObjectIDFromHex(s); err == nil {
			r.id = oid
		}
	}
}

func (r *genericRepository[T]) RegisterHandle(name string, fn func(ctx context.Context, data any, prefix string) types.Message) {
	registerHandle(r.modelName, name, fn)
}

func (r *genericRepository[T]) RegisterModel() {
	mongox.Register(r.NewEntity(), true)
}

func (r *genericRepository[T]) CallFunc(name string, data any) (message types.Message) {
	return callFunc(context.Background(), r.modelName, r.prefix, name, data)
}

func (r *genericRepository[T]) Create(data T) (message types.Message) {
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
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
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

func (r *genericRepository[T]) CreateMany(data []T) (message types.Message) {
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
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
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

func (r *genericRepository[T]) Update(data T) (message types.Message) {
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
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
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

func (r *genericRepository[T]) Delete(delete_message types.DeletePost) (message types.Message) {
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &delete_message); err.HasError() {
		return err
	}

	if delete_message.AfterApprove {
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterApprove, &delete_message); err.HasError() {
			return err
		}
	}
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
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

func (r *genericRepository[T]) DeleteBy(data T) (message types.Message) {
	switch db := r.db.(type) {
	case *mongo.Database:
		var err error
		var delete_message types.DeletePost

		if hook, ok := any(&data).(hooks.HasDeletePost); ok {
			delete_message = hook.GetDeletePost()
			if delete_message.Id.IsZero() {
				message.Status = "error"
				message.Code = 400
				message.Message = "Data does not exists"
				return
			}
		}

		if hook, ok := any(&data).(hooks.BeforeDeleteHook); ok {
			if err := hook.BeforeDelete(); err != nil {
				message.Status = "error"
				message.Code = 400
				message.Message = "Data update error, please contact adminstrator"
				message.MessageErr = err
				return
			}
		}

		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &data); err.HasError() {
			message.Status = "error"
			message.Code = 400
			message.MessageErr = err.MessageErr
			return
		}

		id, err := primitive.ObjectIDFromHex(string(delete_message.Id))
		if err != nil {
			message.Status = "error"
			message.Code = 400
			message.Message = "Data create error, please contact adminstrator"
			return
		}

		_, err = db.Collection(r.collectionName).UpdateOne(
			context.Background(),
			bson.M{"_id": id},
			bson.M{
				"$set": bson.M{
					"deleted_at":    time.Now().UTC(),
					"delete_reason": delete_message.DeleteReason,
					"is_deleted":    true,
				},
			},
		)
		if err != nil {
			message.Status = "error"
			message.Code = 400
			message.Message = "Delete error, please contact adminstrator"
			return
		}

		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterDelete, &data); err.HasError() {
			message.Status = "error"
			message.Code = 400
			message.MessageErr = err.MessageErr
			return
		}

		message.Status = "success"
		message.Code = 200
		message.Message = "Delete successfuly!"
	case *gorm.DB:
		_ = db
	}
	return
}

func (r *genericRepository[T]) GetByFilter(filterStr string) (message types.Message) {
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
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
	return types.Message{
		Status: "success",
		Code:   200,
		Data:   r.response,
	}
}

func (r *genericRepository[T]) ListPage(filterStr string) (message types.Message) {
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

	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
	return types.Message{
		Status:    "success",
		Code:      200,
		Data:      r.response,
		TotalRow:  100,
		TotalPage: int64(math.Ceil(float64(100) / float64(opts_filter.Limit()))),
	}
}

func (r *genericRepository[T]) Aggregate(filter []bson.M) (message types.Message) {
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
	return
}

func (r *genericRepository[T]) UpdateMany(filterStr string, data map[string]any) (message types.Message) {
	switch db := r.db.(type) {
	case *mongo.Database:
		_ = db
	case *gorm.DB:
		_ = db
	}
	return
}
