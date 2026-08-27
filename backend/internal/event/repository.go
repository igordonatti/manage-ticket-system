package event

import (
	"context"
	"errors"
)

var (
	ErrEventNotFound      = errors.New("event not found")
	ErrEventAlreadyExists = errors.New("event already exists")
)

type Repository interface {
	Create(ctx context.Context, event *Event) error
	FindById(ctx context.Context, id string) (*Event, error)
}
