package mongo

import (
	"context"
	"go/types"
	"math"

	// "xp1/config"
	// "xp1/util"

	hooks "github.com/stingwolf1080/dynamic-filter/hook"
	"github.com/stingwolf1080/dynamic-filter/pkg/helper"
	"github.com/stingwolf1080/dynamic-filter/pkg/mongox"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type repository[T any] struct {
	modelName string
	respone   interface{}
	prefix    string
	id        primitive.ObjectID
	restrict  map[string]bool
	alias     map[string]string
	isReturn  bool
	timezone  string
}

func NewRepository[T any](restrict map[string]bool, alias map[string]string) Repository[T] {
	modelName := helper.GetNameModel[T]()
	r := &repository[T]{
		modelName: modelName,
		restrict:  restrict,
		alias:     alias,
	}
	RegisterDeleteFunc(modelName, r)
	return r
}

func (r *repository[T]) RegisterModel() {
	mongox.Register(r.NewEntity(), true)
}

func (r *repository[T]) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *repository[T]) SetID(id primitive.ObjectID) {
	r.id = id
}

func (r *repository[T]) SetRespone(data interface{}) {
	r.respone = data
}

func (r *repository[T]) SetIsReturn() {
	r.isReturn = true
}

func (r *repository[T]) SetTimezone(zone string) {
	r.timezone = zone
}

func (r *repository[T]) CheckFilter(filter string) types.Message {
	var err error
	opts_filter := &util.FilterOptions{RestrictField: r.restrict}
	err = util.CheckParams(filter, opts_filter)
	if err != nil {
		return types.Message{
			Status:  "error",
			Code:    400,
			Message: err.Error(),
		}
	}
	return types.Message{
		Status: "success",
		Code:   200,
	}
}

func (r *repository[T]) RegisterHandle(name string, fn func(ctx context.Context, data any, prefix string) types.Message) {
	registerHandle(r.modelName, name, fn)
}

func (r *repository[T]) CallFunc(name string, data any) types.Message {
	return callFunc(context.Background(), r.modelName, r.prefix, name, data)
}

func (r *repository[T]) NewEntity() T {
	var entity T
	return entity
}

func (r *repository[T]) CreateMany(data []T) types.Message {
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)
	var docs []interface{}
	var ids []primitive.ObjectID
	for k, _data := range data {
		if hook, ok := any(&_data).(hooks.BeforeInsertHook); ok {
			if err := hook.BeforeInsert(); err != nil {
				return types.Message{
					Status:     "error",
					Code:       400,
					Message:    "Data insert error, please contact adminstrator",
					MessageErr: err,
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
				// fmt.Printf("ids: %v\n", id)
				ids = append(ids, util.ObjectID(id.Hex()))
			}
		}
	}

	collecion.Data = docs
	err := collecion.CreateMany()
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Data create error, please contact adminstrator",
			MessageErr: err,
		}
	}
	for k, _data := range data {
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterInsert, &_data); err.HasError() {
			return err
		}
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterSave, &_data); err.HasError() {
			return err
		}
		if hook, ok := any(&_data).(hooks.AfterInsertHook); ok {
			// fmt.Printf("ids: %v\n", ids)
			if err := hook.AfterInsert(ids); err != nil {
				return types.Message{
					Status:     "error",
					Code:       400,
					Message:    "Data insert error, please contact adminstrator",
					MessageErr: err,
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

func (r *repository[T]) Create(data T) types.Message {
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)

	if hook, ok := any(&data).(hooks.BeforeInsertHook); ok {
		if err := hook.BeforeInsert(); err != nil {
			return types.Message{
				Status:     "error",
				Code:       400,
				Message:    "Data insert error, please contact adminstrator",
				MessageErr: err,
			}
		}
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeInsert, &data); err.HasError() {
		return err
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeSave, &data); err.HasError() {
		return err
	}

	collecion.Data = data
	err := collecion.Create()
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Data create error, please contact adminstrator",
			MessageErr: err,
		}
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterInsert, &data); err.HasError() {
		return err
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterSave, &data); err.HasError() {
		return err
	}
	if hook, ok := any(&data).(hooks.HasReturned); ok {
		if hook.HasReturned() {
			return types.Message{
				Status:  "success",
				Code:    201,
				Message: "Create successfuly!",
				Data:    data,
			}
		}
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

func (r *repository[T]) Update(data T) types.Message {
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)

	var r_id primitive.ObjectID
	var none_id bool
	if hook, ok := any(&data).(hooks.HasID); ok {
		none_id, r_id = hook.GetID()
		if !none_id {
			if collecion.CheckObjectIdNil(r_id) {
				return types.Message{
					Status:  "error",
					Code:    400,
					Message: "Data does not exists",
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
			}
		}
		if none_id {
			if hook, ok := any(&data).(hooks.HasID); ok {
				_, r_id = hook.GetID()
				if collecion.CheckObjectIdNil(r_id) {
					return types.Message{
						Status:  "error",
						Code:    400,
						Message: "Data does not exists",
					}
				}
			}
		}
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeSave, &data); err.HasError() {
		return err
	}
	collecion.Data = data
	err := collecion.UpdateId(r_id)
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Data update error, please contact adminstrator",
			MessageErr: err,
		}
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterUpdate, &data); err.HasError() {
		return err
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterSave, &data); err.HasError() {
		return err
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Update successfuly!",
	}
}
func (r *repository[T]) Delete(delete_message util.DeletePost) types.Message {
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)
	// var err error
	// if hook, ok := any(&delete_message).(hooks.DefaultHook); ok {
	// if err := hook.BeforeDelete(); err.HasError() {
	// return err
	// }
	// }
	// if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &data); err.HasError() {
	// return err
	// }
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &delete_message); err.HasError() {
		return err
	}
	if delete_message.AfterApprove {
		if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterApprove, &delete_message); err.HasError() {
			return err
		}
	} else {
		id, err := delete_message.Id.ObjectID()
		if err != nil {
			return types.Message{
				Status:  "error",
				Code:    400,
				Message: "Data create error, please contact adminstrator",
			}
		}
		// log.Println("block soft deleted")
		err = collecion.SoftDeleteId(id, delete_message.DeleteReason)
		if err != nil {
			return types.Message{
				Status:  "error",
				Code:    400,
				Message: "Data create error, please contact adminstrator",
			}
		}
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Delete successfuly!",
	}
}

func (r *repository[T]) GetByFilter(filter string) types.Message {
	var err error
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)
	data := r.NewEntity()
	collecion.Data = &data
	if r.respone != nil {
		collecion.Data = r.respone
	}
	opts_filter := &util.FilterOptions{AliasField: r.alias}
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}
	err = util.ParseBracketParams(filter, opts_filter)
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Unable to parse: an object hierarchy has been provided",
			MessageErr: err,
		}
	}
	collecion.Filter = opts_filter
	err = collecion.Read()
	if err != nil {
		if collecion.CheckNotFound(err) {
			return types.Message{
				Status:  "error",
				Code:    404,
				Message: "Data not found or deactive",
			}
		}
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Load data error, please contact adminstrator",
			MessageErr: err,
		}
	}
	return types.Message{
		Status: "success",
		Code:   200,
		Data:   r.respone,
	}
}

func (r *repository[T]) ListPage(filter string) types.Message {
	// log.Println(filter)
	var data []T
	var err error
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)
	collecion.Data = &data
	// if r.respone != nil {
	// 	collecion.Data = r.respone
	// }
	opts_filter := &util.FilterOptions{AliasField: r.alias}
	opts_filter.SetPagination()
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}
	err = util.ParseBracketParams(filter, opts_filter)
	if err != nil {
		return types.Message{
			Status:  "error",
			Code:    400,
			Message: err.Error(),
		}
	}
	collecion.Filter = opts_filter
	if len(opts_filter.GroupBy) > 0 {
		collecion.Data = r.NewEntity()
		err = collecion.ReadGroupBy(r.prefix)
		if err != nil {
			return types.Message{
				Status:     "error",
				Code:       400,
				MessageErr: err,
				Message:    "Load data error, please contact adminstrator",
			}
		}
		return types.Message{
			Status: "success",
			Code:   200,
			Data:   collecion.Data,
		}
	}
	err = collecion.CountDocuments()
	if collecion.CheckNotFound(err) {
		return types.Message{
			Status:  "success",
			Code:    200,
			Message: "Data does not exists",
		}
	}
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Load data error, please contact adminstrator",
			MessageErr: err,
		}
	}
	err = collecion.ReadWithOption()
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Load data error, please contact adminstrator",
			MessageErr: err,
		}
	}
	return types.Message{
		Status:    "success",
		Code:      200,
		Data:      data,
		TotalRow:  collecion.Count,
		TotalPage: int64(math.Ceil(float64(collecion.Count) / float64(opts_filter.Limit()))),
	}
}

func (r *repository[T]) UpdateMany(filter string, data map[string]any) util.ErrorMessage {
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)

	if len(data) <= 0 {
		return util.ErrorMessage{
			Status:  "error",
			Code:    400,
			Message: "Data update error, please contact adminstrator",
		}
	}
	collecion.Data = data

	opts_filter := &util.FilterOptions{AliasField: r.alias}
	if r.timezone != "" {
		opts_filter.SetTimezone(r.timezone)
	}
	var err error
	err = util.ParseBracketParams(filter, opts_filter)
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Unable to parse: an object hierarchy has been provided",
			MessageErr: err,
		}
	}
	collecion.Filter = opts_filter

	err = collecion.UpdateMany()
	if err != nil {
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Data update error, please contact adminstrator",
			MessageErr: err,
		}
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Update successfuly!",
	}
}

func (r *repository[T]) DeleteBy(data T) types.Message {
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)
	var err error
	var delete_message util.DeletePost
	if hook, ok := any(&data).(hooks.HasDeletePost); ok {
		delete_message = hook.GetDeletePost()
		if delete_message.Id.IsZero() {
			return types.Message{
				Status:  "error",
				Code:    400,
				Message: "Data does not exists",
			}
		}
	}

	if hook, ok := any(&data).(hooks.BeforeDeleteHook); ok {
		if err := hook.BeforeDelete(); err != nil {
			return types.Message{
				Status:     "error",
				Code:       400,
				Message:    "Data update error, please contact adminstrator",
				MessageErr: err,
			}
		}
	}

	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.BeforeDelete, &data); err.HasError() {
		return err
	}
	id, err := delete_message.Id.ObjectID()
	if err != nil {
		return types.Message{
			Status:  "error",
			Code:    400,
			Message: "Data create error, please contact adminstrator",
		}
	}
	err = collecion.SoftDeleteId(id, delete_message.DeleteReason)
	if err != nil {
		return types.Message{
			Status:  "error",
			Code:    400,
			Message: "Delete error, please contact adminstrator",
		}
	}
	if err := hooks.Run(context.Background(), r.modelName, r.prefix, hooks.AfterDelete, &data); err.HasError() {
		return err
	}
	return types.Message{
		Status:  "success",
		Code:    200,
		Message: "Delete successfuly!",
	}
}

func (r *repository[T]) Aggregate(filter []bson.M) types.Message {
	var err error
	collecion := new(config.Collection)
	collecion.SetTable(r.NewEntity(), r.prefix)
	data := r.NewEntity()
	collecion.Data = &data
	if r.respone != nil {
		collecion.Data = r.respone
	}
	err = collecion.Aggregate(filter)
	if err != nil {
		if collecion.CheckNotFound(err) {
			return types.Message{
				Status:  "error",
				Code:    404,
				Message: "Data not found or deactive",
			}
		}
		return types.Message{
			Status:     "error",
			Code:       400,
			Message:    "Load data error, please contact adminstrator",
			MessageErr: err,
		}
	}
	return types.Message{
		Status: "success",
		Code:   200,
		Data:   r.respone,
	}
}
