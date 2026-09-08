package redisstore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/igordonatti/sistema-gestao-ingressos/backend/internal/event"
)

func TestEventRepositoryCreateFindAndReopen(t *testing.T) {
	if os.Getenv("REDIS_INTEGRATION") != "1" {
		t.Skip("set REDIS_INTEGRATION=1 to run Redis integration tests")
	}

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := New(addr)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		t.Fatalf("Redis is not available at %s: %v", addr, err)
	}

	repository := NewEventRepository(client)
	eventID := fmt.Sprintf(
		"integration-%d",
		time.Now().UnixNano(),
	)

	t.Cleanup(func() {
		cleanupEventRepositoryTest(t, addr, eventID)
	})

	eventStart := time.Date(
		2026,
		time.October,
		17,
		23,
		0,
		0,
		0,
		time.UTC,
	)
	salesStart := time.Date(
		2026,
		time.September,
		1,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	createdEvent, err := event.New(
		eventID,
		"Evento de integração",
		"Local de integração",
		eventStart,
		salesStart,
		100,
		5000,
	)
	if err != nil {
		t.Fatalf("create domain event: %v", err)
	}

	if err := repository.Create(ctx, createdEvent); err != nil {
		t.Fatalf("persist event: %v", err)
	}

	if err := repository.Create(ctx, createdEvent); !errors.Is(
		err,
		event.ErrEventAlreadyExists,
	) {
		t.Fatalf(
			"second create expected ErrEventAlreadyExists, received %v",
			err,
		)
	}

	loadedEvent, err := repository.FindByID(ctx, eventID)
	if err != nil {
		t.Fatalf("find persisted event: %v", err)
	}

	if loadedEvent.ID() != createdEvent.ID() {
		t.Errorf("ID expected %s, received %s", createdEvent.ID(), loadedEvent.ID())
	}

	if loadedEvent.Name() != createdEvent.Name() {
		t.Errorf("name expected %s, received %s", createdEvent.Name(), loadedEvent.Name())
	}

	if !loadedEvent.EventStart().Equal(createdEvent.EventStart()) {
		t.Errorf("event start expected %s, received %s", createdEvent.EventStart(), loadedEvent.EventStart())
	}

	availability, err := repository.FindAvailabilityByEventID(ctx, eventID)
	if err != nil {
		t.Fatalf("find persisted availability: %v", err)
	}

	if availability.Capacity() != 100 || availability.Available() != 100 {
		t.Errorf(
			"initial availability expected capacity=100 and available=100, received capacity=%d available=%d",
			availability.Capacity(),
			availability.Available(),
		)
	}

	if err := client.Close(); err != nil {
		t.Fatalf("close first Redis client: %v", err)
	}

	reopenedClient := New(addr)
	defer reopenedClient.Close()
	reopenedRepository := NewEventRepository(reopenedClient)

	reopenedEvent, err := reopenedRepository.FindByID(ctx, eventID)
	if err != nil {
		t.Fatalf("find event with reopened client: %v", err)
	}

	if reopenedEvent.ID() != eventID {
		t.Errorf("reopened event ID expected %s, received %s", eventID, reopenedEvent.ID())
	}

	_, err = reopenedRepository.FindByID(ctx, "does-not-exist")
	if !errors.Is(err, event.ErrEventNotFound) {
		t.Fatalf("missing event expected ErrEventNotFound, received %v", err)
	}
}

func cleanupEventRepositoryTest(t *testing.T, addr string, eventID string) {
	t.Helper()

	cleanupClient := New(addr)
	defer cleanupClient.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Second,
	)
	defer cancel()

	cleanupClient.client.Del(
		ctx,
		eventKey(eventID),
		inventoryKey(eventID),
	)
	cleanupClient.client.ZRem(ctx, eventsByStartKey, eventID)
}
