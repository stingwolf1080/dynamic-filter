package hooks

import (
	"github.com/stingwolf1080/dynamic-filter/pkg/util/types"
)

type HasID interface {
	GetID() (bool, types.ID)
}

type HasReturned interface {
	HasReturned() bool
}

type HasDeletePost interface {
	GetDeletePost() types.DeletePost
}

type BeforeInsertHook interface {
	BeforeInsert() error
}

type BeforeUpdateHook interface {
	BeforeUpdate() error
}

type BeforeDeleteHook interface {
	BeforeDelete() error
}

type AfterInsertHook interface {
	AfterInsert(ids []types.ID) error
}
