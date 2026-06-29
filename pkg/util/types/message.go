package types

import (
	"reflect"
)

func IsPointerAnArray(value any) bool {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		return false
	}
	elem := v.Elem()
	if elem.Kind() == reflect.Array || elem.Kind() == reflect.Slice {
		return true
	}
	return false
}

type Message struct {
	Status     string `json:"status"`
	Message    string `json:"message,omitempty"`
	Code       int    `json:"code"`
	LastId     string `json:"next,omitempty"`
	FirstId    string `json:"prev,omitempty"`
	Data       any    `json:"data,omitempty"`
	MessageErr error  `json:"-"`
	ErrorCode  int    `json:"error_code,omitempty"`
	TotalRow   int64  `json:"total_row,omitempty"`
	TotalPage  int64  `json:"total_page,omitempty"`
}

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

type ResponeMessage[T any] struct {
	Status    string `json:"status" example:"success"`
	Message   string `json:"message,omitempty"`
	Code      int    `json:"code" swaggertype:"number" example:"200"`
	Data      T      `json:"data"`
	TotalRow  int64  `json:"total_row" swaggertype:"number" example:"1"`
	TotalPage int64  `json:"total_page" swaggertype:"number" example:"10"`
}

type ObjectResponse[T any] struct {
	Status  string `json:"status"  example:"success"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code" swaggertype:"number" example:"200"`
	Data    T      `json:"data"`
}

type SuccessRespone struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message"`
	Code    int    `json:"code" swaggertype:"number" example:"200"`
}

type BadRequestResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message,omitempty" example:"Bad request"`
	Code    int    `json:"code" swaggertype:"number" example:"400"`
}

type NotFoundResponse struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Data not found"`
	Code    int    `json:"code" swaggertype:"number" example:"404"`
}

type PermisionRespone struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Permision denied"`
	Code    int    `json:"code" swaggertype:"number" example:"403"`
}

type UnauthorizedRespone struct {
	Status  string `json:"status" example:"error"`
	Message string `json:"message" example:"Unauthorized"`
	Code    int    `json:"code" swaggertype:"number" example:"401"`
}

type CreateHasRespone[T any] struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message"`
	Code    int    `json:"code" swaggertype:"number" example:"200"`
	Data    T      `json:"data"`
}

type CreateSuccessRespone struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Create successfully!"`
	Code    int    `json:"code" swaggertype:"number" example:"201"`
}

type UpdateSuccessRespone struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Update successfully!"`
	Code    int    `json:"code" swaggertype:"number" example:"200"`
}

type DeleteSuccessRespone struct {
	Status  string `json:"status" example:"success"`
	Message string `json:"message" example:"Delete successfully!"`
	Code    int    `json:"code" swaggertype:"number" example:"200"`
}

func (e *Message) HasError() bool {
	return e.Code == 400
}

func (e *Message) HasNotFound() bool {
	return e.Code == 404
}

func (e *Message) GetStatus() string {
	return e.Status
}

func (e *Message) GetMessage() string {
	return e.Message
}

func (e *Message) CheckData(data any) bool {
	if reflect.TypeOf(e.Data) == reflect.TypeOf(data) {
		return true
	}
	return false
}
