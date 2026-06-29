package types

type DeletePost struct {
	Id           ObjectID `json:"id" validate:"required"`
	DeleteReason string   `json:"reason" validate:"required"`
	AfterApprove bool     `json:"-"`
	Owner        ObjectID `json:"-"`
}
