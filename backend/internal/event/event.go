package event

// name string // privado
// Name string // exportado
import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

type Status string

const StatusDraft Status = "draft"

type Event struct {
	id         string
	name       string
	location   string
	eventStart time.Time
	salesStart time.Time
	capacity   int
	priceCents int64
	status     Status
}

var (
	ErrInvalidID         = errors.New("event ID cannot be empty")
	ErrInvalidName       = errors.New("event name must have at least 3 characters")
	ErrInvalidLocation   = errors.New("event location must have at least 3 characters")
	ErrInvalidCapacity   = errors.New("event capacity must be greater than zero")
	ErrInvalidPrice      = errors.New("evenr price cannot be negative")
	ErrInvalidSalesStart = errors.New("sales must start before the event")
)

// funciona como um cunstructor da classe
func New(
	id string,
	name string,
	location string,
	eventStart time.Time,
	salesStart time.Time,
	capacity int,
	priceCents int64,
) (*Event, error) {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	location = strings.TrimSpace(location)

	if id == "" {
		return nil, ErrInvalidID
	}

	if utf8.RuneCountInString(name) < 3 {
		return nil, ErrInvalidName
	}

	if utf8.RuneCountInString(location) < 3 {
		return nil, ErrInvalidLocation
	}

	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	if priceCents < 0 {
		return nil, ErrInvalidPrice
	}

	if !salesStart.Before(eventStart) {
		return nil, ErrInvalidSalesStart
	}

	return &Event{
		id:         id,
		name:       name,
		location:   location,
		eventStart: eventStart,
		salesStart: salesStart,
		capacity:   capacity,
		priceCents: priceCents,
		status:     StatusDraft,
	}, nil
}

func (e *Event) ID() string {
	return e.id
}

func (e *Event) Name() string {
	return e.name
}

func (e *Event) Location() string {
	return e.location
}

func (e *Event) EventStart() time.Time {
	return e.eventStart
}

func (e *Event) SalesStart() time.Time {
	return e.salesStart
}

func (e *Event) Capacity() int {
	return e.capacity
}

func (e *Event) PriceCents() int64 {
	return e.priceCents
}

func (e *Event) Status() Status {
	return e.status
}
