package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" binding:"required,max=100"`
	Description *string   `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
	CreatedAt   time.Time `json:"created_at"`
}

const createTableSQL = `
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT end_after_start CHECK (end_time > start_time)
);
`

func InitSchema(db *sql.DB) error {
	_, err := db.Exec(createTableSQL)
	return err
}

type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (id, title, description, start_time, end_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	event.ID = uuid.New().String()
	event.CreatedAt = time.Now().UTC()

	err := r.db.QueryRowContext(
		ctx,
		query,
		event.ID,
		event.Title,
		event.Description,
		event.StartTime.UTC(),
		event.EndTime.UTC(),
		event.CreatedAt,
	).Scan(&event.ID, &event.CreatedAt)

	return err
}

func (r *EventRepository) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, created_at
		FROM events
		WHERE id = $1
	`

	event := &Event{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.StartTime,
		&event.EndTime,
		&event.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) List(ctx context.Context) ([]*Event, error) {
	query := `
		SELECT id, title, description, start_time, end_time, created_at
		FROM events
		ORDER BY start_time ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		if err := rows.Scan(
			&event.ID,
			&event.Title,
			&event.Description,
			&event.StartTime,
			&event.EndTime,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
