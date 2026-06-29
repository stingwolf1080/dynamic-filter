package types

type CancelPost struct {
	Id           ObjectID `json:"id" validate:"required"`
	CancelReason string   `json:"reason" validate:"required"`
}
