package entities

import (
	"github.com/pkg/errors"
)

var (
	ErrInvalidParam          = errors.New("invalid param")
	ErrInvalidScript         = errors.New("invalid script")
	ErrInvalidAggregateParam = errors.New("invalid aggregate param")
	ErrInternal              = errors.New("internal error")
	ErrNotFound              = errors.New("not found")
)
