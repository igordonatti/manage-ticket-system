package event_test

import (
	"errors"
	"testing"
	"time"

	"github.com/igordonatti/sistema-gestao-ingressos/backend/internal/event"
)

type eventInput struct {
	id         string
	name       string
	location   string
	eventStart time.Time
	salesStart time.Time
	capacity   int
	priceCents int64
}

func validEventInput() eventInput {
	return eventInput{
		id:         "event-001",
		name:       "Órbita Sonora",
		location:   "Galpão 67",
		eventStart: time.Date(2026, time.October, 17, 23, 0, 0, 0, time.UTC),
		salesStart: time.Date(2026, time.September, 1, 10, 0, 0, 0, time.UTC),
		capacity:   2_400,
		priceCents: 14_500,
	}
}

func createEvent(input eventInput) (*event.Event, error) {
	return event.New(
		input.id,
		input.name,
		input.location,
		input.eventStart,
		input.salesStart,
		input.capacity,
		input.priceCents,
	)
}

func TestNewCreatesValidDraftEvent(t *testing.T) {
	input := validEventInput()
	input.id = "  event-001  "
	input.name = "  Órbita Sonora  "
	input.location = "  Galpão 67  "

	got, err := createEvent(input)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	if got == nil {
		t.Fatal("New() returned a nil event without an error")
	}

	if got.ID() != "event-001" {
		t.Errorf("ID() = %q; want %q", got.ID(), "event-001")
	}
	if got.Name() != "Órbita Sonora" {
		t.Errorf("Name() = %q; want %q", got.Name(), "Órbita Sonora")
	}
	if got.Location() != "Galpão 67" {
		t.Errorf("Location() = %q; want %q", got.Location(), "Galpão 67")
	}
	if !got.EventStart().Equal(input.eventStart) {
		t.Errorf("EventStart() = %v; want %v", got.EventStart(), input.eventStart)
	}
	if !got.SalesStart().Equal(input.salesStart) {
		t.Errorf("SalesStart() = %v; want %v", got.SalesStart(), input.salesStart)
	}
	if got.Capacity() != input.capacity {
		t.Errorf("Capacity() = %d; want %d", got.Capacity(), input.capacity)
	}
	if got.PriceCents() != input.priceCents {
		t.Errorf("PriceCents() = %d; want %d", got.PriceCents(), input.priceCents)
	}
	if got.Status() != event.StatusDraft {
		t.Errorf("Status() = %q; want %q", got.Status(), event.StatusDraft)
	}
}

func TestNewAllowsFreeEvent(t *testing.T) {
	input := validEventInput()
	input.priceCents = 0

	got, err := createEvent(input)
	if err != nil {
		t.Fatalf("New() returned unexpected error for a free event: %v", err)
	}

	if got.PriceCents() != 0 {
		t.Errorf("PriceCents() = %d; want 0", got.PriceCents())
	}
}

func TestNewRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*eventInput)
		wantErr error
	}{
		{
			name:    "empty id",
			change:  func(input *eventInput) { input.id = "   " },
			wantErr: event.ErrInvalidID,
		},
		{
			name:    "empty name",
			change:  func(input *eventInput) { input.name = "   " },
			wantErr: event.ErrInvalidName,
		},
		{
			name:    "name shorter than three ASCII characters",
			change:  func(input *eventInput) { input.name = "AB" },
			wantErr: event.ErrInvalidName,
		},
		{
			name:    "name shorter than three Unicode characters",
			change:  func(input *eventInput) { input.name = "Ór" },
			wantErr: event.ErrInvalidName,
		},
		{
			name:    "empty location",
			change:  func(input *eventInput) { input.location = "   " },
			wantErr: event.ErrInvalidLocation,
		},
		{
			name:    "location shorter than three ASCII characters",
			change:  func(input *eventInput) { input.location = "AB" },
			wantErr: event.ErrInvalidLocation,
		},
		{
			name:    "location shorter than three Unicode characters",
			change:  func(input *eventInput) { input.location = "Sé" },
			wantErr: event.ErrInvalidLocation,
		},
		{
			name:    "zero capacity",
			change:  func(input *eventInput) { input.capacity = 0 },
			wantErr: event.ErrInvalidCapacity,
		},
		{
			name:    "negative capacity",
			change:  func(input *eventInput) { input.capacity = -1 },
			wantErr: event.ErrInvalidCapacity,
		},
		{
			name:    "negative price",
			change:  func(input *eventInput) { input.priceCents = -1 },
			wantErr: event.ErrInvalidPrice,
		},
		{
			name: "sales starting at event time",
			change: func(input *eventInput) {
				input.salesStart = input.eventStart
			},
			wantErr: event.ErrInvalidSalesStart,
		},
		{
			name: "sales starting after event",
			change: func(input *eventInput) {
				input.salesStart = input.eventStart.Add(time.Second)
			},
			wantErr: event.ErrInvalidSalesStart,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validEventInput()
			test.change(&input)

			got, err := createEvent(input)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("New() error = %v; want %v", err, test.wantErr)
			}
			if got != nil {
				t.Errorf("New() returned event %#v with invalid input; want nil", got)
			}
		})
	}
}
