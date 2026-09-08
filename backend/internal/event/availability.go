package event

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidAvailability = errors.New(
	"invalid event availability",
)

type Availability struct {
	eventID   string
	capacity  int
	available int
	reserved  int
	sold      int
}

func NewAvailability(
	eventID string,
	capacity int,
	available int,
	reserved int,
	sold int,
) (*Availability, error) {
	eventID = strings.TrimSpace(eventID)

	if eventID == "" {
		return nil, fmt.Errorf(
			"%w: event ID cannot be empty",
			ErrInvalidAvailability,
		)
	}

	if capacity <= 0 {
		return nil, fmt.Errorf(
			"%w: capacity must be greater than zero",
			ErrInvalidAvailability,
		)
	}

	if available < 0 || reserved < 0 || sold < 0 {
		return nil, fmt.Errorf(
			"%w: counters cannot be negative",
			ErrInvalidAvailability,
		)
	}

	if available+reserved+sold != capacity {
		return nil, fmt.Errorf(
			"%w: counters must add up to capacity",
			ErrInvalidAvailability,
		)
	}

	return &Availability{
		eventID:   eventID,
		capacity:  capacity,
		available: available,
		reserved:  reserved,
		sold:      sold,
	}, nil
}

func (a *Availability) EventID() string {
	return a.eventID
}

func (a *Availability) Capacity() int {
	return a.capacity
}

func (a *Availability) Available() int {
	return a.available
}

func (a *Availability) Reserved() int {
	return a.reserved
}

func (a *Availability) Sold() int {
	return a.sold
}
