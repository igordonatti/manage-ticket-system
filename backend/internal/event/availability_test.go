package event

import (
	"errors"
	"testing"
)

func TestNewAvailabilityValid(t *testing.T) {
	availability, err := NewAvailability(
		"  event-001  ",
		100,
		70,
		10,
		20,
	)
	if err != nil {
		t.Fatalf("expected no error, received %v", err)
	}

	if availability.EventID() != "event-001" {
		t.Errorf("event ID expected event-001, received %s", availability.EventID())
	}

	if availability.Capacity() != 100 {
		t.Errorf("capacity expected 100, received %d", availability.Capacity())
	}

	if availability.Available() != 70 {
		t.Errorf("available expected 70, received %d", availability.Available())
	}

	if availability.Reserved() != 10 {
		t.Errorf("reserved expected 10, received %d", availability.Reserved())
	}

	if availability.Sold() != 20 {
		t.Errorf("sold expected 20, received %d", availability.Sold())
	}
}

func TestNewAvailabilityInvalid(t *testing.T) {
	tests := []struct {
		name      string
		eventID   string
		capacity  int
		available int
		reserved  int
		sold      int
	}{
		{
			name:     "empty event ID",
			eventID:  "   ",
			capacity: 100,
		},
		{
			name:     "zero capacity",
			eventID:  "event-001",
			capacity: 0,
		},
		{
			name:      "negative available",
			eventID:   "event-001",
			capacity:  100,
			available: -1,
			reserved:  50,
			sold:      51,
		},
		{
			name:      "negative reserved",
			eventID:   "event-001",
			capacity:  100,
			available: 50,
			reserved:  -1,
			sold:      51,
		},
		{
			name:      "negative sold",
			eventID:   "event-001",
			capacity:  100,
			available: 50,
			reserved:  51,
			sold:      -1,
		},
		{
			name:      "counters do not add up",
			eventID:   "event-001",
			capacity:  100,
			available: 50,
			reserved:  20,
			sold:      20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAvailability(
				tt.eventID,
				tt.capacity,
				tt.available,
				tt.reserved,
				tt.sold,
			)

			if !errors.Is(err, ErrInvalidAvailability) {
				t.Fatalf(
					"expected ErrInvalidAvailability, received %v",
					err,
				)
			}
		})
	}
}
