package redisstore

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/igordonatti/sistema-gestao-ingressos/backend/internal/event"
	"github.com/redis/go-redis/v9"
)

const (
	eventsByStartKey      = "events:by_start"
	createEventMaxRetries = 3
)

type EventRepository struct {
	client *Client
}

// NO MOMENTO DE COMPILAR: EventRepository implementa todos os métodos de event.Repository?
var _ event.Repository = (*EventRepository)(nil)
var _ event.AvailabilityReader = (*EventRepository)(nil)

func NewEventRepository(client *Client) *EventRepository {
	return &EventRepository{
		client: client,
	}
}

func (r *EventRepository) Create(
	ctx context.Context,
	eventToCreate *event.Event,
) error {
	key := eventKey(eventToCreate.ID())

	for attempt := 0; attempt < createEventMaxRetries; attempt++ {
		err := r.client.client.Watch(
			ctx,
			func(tx *redis.Tx) error {
				exists, err := tx.Exists(ctx, key).Result()
				if err != nil {
					return fmt.Errorf("check event existence: %w", err)
				}

				if exists > 0 {
					return event.ErrEventAlreadyExists
				}

				_, err = tx.TxPipelined(
					ctx,
					func(pipe redis.Pipeliner) error {
						pipe.HSet(
							ctx,
							key,
							eventToHashFields(eventToCreate),
						)

						pipe.HSet(
							ctx,
							inventoryKey(eventToCreate.ID()),
							inventoryToHashFields(eventToCreate),
						)

						pipe.ZAdd(
							ctx,
							eventsByStartKey,
							redis.Z{
								Score: float64(
									eventToCreate.EventStart().UnixMilli(),
								),
								Member: eventToCreate.ID(),
							},
						)

						return nil
					},
				)
				return err
			},
			key,
		)
		switch {
		case err == nil:
			return nil

		case errors.Is(err, event.ErrEventAlreadyExists):
			return event.ErrEventAlreadyExists

		case errors.Is(err, redis.TxFailedErr):
			continue

		default:
			return fmt.Errorf("create event: %w", err)
		}
	}

	return event.ErrEventAlreadyExists
}

func (r *EventRepository) FindByID(
	ctx context.Context,
	id string,
) (*event.Event, error) {
	fields, err := r.client.client.HGetAll(
		ctx,
		eventKey(id),
	).Result()
	if err != nil {
		return nil, fmt.Errorf(
			"find event %q: %w",
			id,
			err,
		)
	}

	if len(fields) == 0 {
		return nil, event.ErrEventNotFound
	}

	storedEvent, err := eventFromHashFields(fields)
	if err != nil {
		return nil, fmt.Errorf(
			"decode event %q: %w",
			id,
			err,
		)
	}

	if storedEvent.ID() != id {
		return nil, fmt.Errorf(
			"event key for %q contains ID %q",
			id,
			storedEvent.ID(),
		)
	}

	return storedEvent, nil
}

func (r *EventRepository) FindAvailabilityByEventID(
	ctx context.Context,
	eventID string,
) (*event.Availability, error) {
	fields, err := r.client.client.HGetAll(
		ctx,
		inventoryKey(eventID),
	).Result()
	if err != nil {
		return nil, fmt.Errorf(
			"find availability for event %q: %w",
			eventID,
			err,
		)
	}

	if len(fields) == 0 {
		return nil, event.ErrAvailabilityNotFound
	}

	capacity, err := parseInventoryField(fields, "capacity")
	if err != nil {
		return nil, fmt.Errorf(
			"decode availability for event %q: %w",
			eventID,
			err,
		)
	}

	available, err := parseInventoryField(fields, "available")
	if err != nil {
		return nil, fmt.Errorf(
			"decode availability for event %q: %w",
			eventID,
			err,
		)
	}

	reserved, err := parseInventoryField(fields, "reserved")
	if err != nil {
		return nil, fmt.Errorf(
			"decode availability for event %q: %w",
			eventID,
			err,
		)
	}

	sold, err := parseInventoryField(fields, "sold")
	if err != nil {
		return nil, fmt.Errorf(
			"decode availability for event %q: %w",
			eventID,
			err,
		)
	}

	availability, err := event.NewAvailability(
		eventID,
		capacity,
		available,
		reserved,
		sold,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"validate availability for event %q: %w",
			eventID,
			err,
		)
	}

	return availability, nil
}

func eventKey(id string) string {
	return "event:{" + id + "}"
}

func inventoryKey(id string) string {
	return "inventory:{" + id + "}"
}

func eventToHashFields(event *event.Event) map[string]string {
	return map[string]string{
		"id":          event.ID(),
		"name":        event.Name(),
		"location":    event.Location(),
		"event_start": event.EventStart().UTC().Format(time.RFC3339Nano),
		"sales_start": event.SalesStart().UTC().Format(time.RFC3339Nano),
		"capacity":    strconv.Itoa(event.Capacity()),
		"price_cents": strconv.FormatInt(event.PriceCents(), 10),
		"status":      string(event.Status()),
	}
}

func eventFromHashFields(
	fields map[string]string,
) (*event.Event, error) {
	requiredFields := []string{
		"id",
		"name",
		"location",
		"event_start",
		"sales_start",
		"capacity",
		"price_cents",
		"status",
	}

	for _, field := range requiredFields {
		if _, exists := fields[field]; !exists {
			return nil, fmt.Errorf(
				"missing event hash field %q",
				field,
			)
		}
	}

	eventStart, err := time.Parse(
		time.RFC3339Nano,
		fields["event_start"],
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse event_start: %w",
			err,
		)
	}

	salesStart, err := time.Parse(
		time.RFC3339Nano,
		fields["sales_start"],
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse sales_start: %w",
			err,
		)
	}

	capacity, err := strconv.Atoi(fields["capacity"])
	if err != nil {
		return nil, fmt.Errorf(
			"parse capacity: %w",
			err,
		)
	}

	priceCents, err := strconv.ParseInt(
		fields["price_cents"],
		10,
		64,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"parse price_cents: %w",
			err,
		)
	}

	status := event.Status(fields["status"])
	if status != event.StatusDraft {
		return nil, fmt.Errorf(
			"unsupported event status %q",
			status,
		)
	}

	storedEvent, err := event.New(
		fields["id"],
		fields["name"],
		fields["location"],
		eventStart,
		salesStart,
		capacity,
		priceCents,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"validate stored event: %w",
			err,
		)
	}

	return storedEvent, nil
}

func inventoryToHashFields(event *event.Event) map[string]string {
	capacity := strconv.Itoa(event.Capacity())

	return map[string]string{
		"capacity":  capacity,
		"available": capacity,
		"reserved":  "0",
		"sold":      "0",
	}
}

func parseInventoryField(
	fields map[string]string,
	name string,
) (int, error) {
	value, exists := fields[name]
	if !exists {
		return 0, fmt.Errorf(
			"missing inventory field %q",
			name,
		)
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf(
			"parse inventory field %q: %w",
			name,
			err,
		)
	}

	return parsed, nil
}
