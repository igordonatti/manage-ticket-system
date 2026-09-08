package event

import (
	"context"
	"errors"
)

var (
	ErrEventNotFound        = errors.New("event not found")
	ErrEventAlreadyExists   = errors.New("event already exists")
	ErrAvailabilityNotFound = errors.New("event availability not found")
)

type Repository interface {
	Create(ctx context.Context, event *Event) error
	FindByID(ctx context.Context, id string) (*Event, error)
}

type AvailabilityReader interface {
	FindAvailabilityByEventID(
		ctx context.Context,
		eventID string,
	) (*Availability, error)
}
