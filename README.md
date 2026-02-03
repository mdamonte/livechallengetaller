# Events API (Go + PostgreSQL)

A small, idiomatic Go REST API to manage Events with PostgreSQL. It demonstrates clean HTTP handlers, input validation, JSON encoding/decoding, context-aware DB calls, and graceful shutdown.

## Features

- **Create Event**: POST `/api/v1/events`
- **List Events**: GET `/api/v1/events`
- **Get Event by ID**: GET `/api/v1/events/{id}`
- Ordered results by `start_time` ascending
- Validation: non-empty title (<= 100 chars), `start_time < end_time`
- Uses `database/sql` with `pgx` driver
- Context for all requests and DB queries

## Tech Stack

- Go 1.21+
- `net/http`, `encoding/json`, `context`, `database/sql`
- Router: `github.com/gin-gonic/gin`
- PostgreSQL driver: `github.com/jackc/pgx/v5/stdlib`
- UUID: `github.com/google/uuid`
- Env loader: `github.com/joho/godotenv`

## Project Structure

```
.
├── handlers/           # HTTP handlers (Create/List/Get)
│   └── event_handlers.go
├── middleware/         # Error handling middleware
│   └── error_handler.go
├── models/             # Event model + repository + schema init
│   └── event.go
├── main.go             # App entrypoint, router, graceful shutdown
├── go.mod              # Go modules
└── README.md           # This file
```

## Database Schema

You can let the app create the table automatically on startup, or run this SQL yourself:

```sql
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(100) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT end_after_start CHECK (end_time > start_time)
);
```

Note: `gen_random_uuid()` requires the `pgcrypto` extension. If you don't have it, either:

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

or remove the default and let the app generate UUIDs in code (which it already does).

## Setup

1. **Create a PostgreSQL database**
   ```sql
   CREATE DATABASE events_db;
   ```

2. **Configure environment**
   Create a `.env` file at the project root:
   ```env
   DATABASE_URL=postgres://USER:PASS@localhost:5432/events_db?sslmode=disable
   ```

3. **Install dependencies**
   ```bash
   go mod tidy
   ```

4. **Run the service**
   ```bash
   go run ./main.go
   ```

Server listens on `http://localhost:8080`.

## API

- **Health Check**
  - GET `/health`

- **Create Event**
  - POST `/api/v1/events`
  - Request (RFC3339 timestamps):
    ```json
    {
      "title": "Team Meeting",
      "description": "Weekly sync",
      "start_time": "2026-02-05T10:00:00Z",
      "end_time": "2026-02-05T11:00:00Z"
    }
    ```
  - Responses:
    - 201 Created with the created event JSON
    - 400 Bad Request for validation errors

- **List Events**
  - GET `/api/v1/events`
  - Responses:
    - 200 OK with JSON array (empty array if none)

- **Get Event by ID**
  - GET `/api/v1/events/{uuid}`
  - Responses:
    - 200 OK with JSON event
    - 400 Bad Request if ID is not a valid UUID
    - 404 Not Found if event does not exist

## Curl Examples

- Create
  ```bash
  curl -sS -X POST http://localhost:8080/api/v1/events \
    -H 'Content-Type: application/json' \
    -d '{
      "title": "Demo",
      "description": "Run-through",
      "start_time": "2026-02-05T10:00:00Z",
      "end_time": "2026-02-05T11:00:00Z"
    }' | jq
  ```

- List
  ```bash
  curl -sS http://localhost:8080/api/v1/events | jq
  ```

- Get by ID
  ```bash
  curl -sS http://localhost:8080/api/v1/events/<uuid> | jq
  ```

## Notes

- The service uses a connection pool with sane defaults (`SetMaxOpenConns`, `SetMaxIdleConns`, `SetConnMaxLifetime`).
- All times are handled/stored in UTC.
- Input is validated: title required and <= 100 chars; `end_time` must be after `start_time`.

## Troubleshooting

- "DATABASE_URL is not set": create `.env` or export the variable in your shell.
- `pq: function gen_random_uuid() does not exist`: enable `pgcrypto` or remove the default from the schema.
- Connection errors: verify host/port/credentials in `DATABASE_URL`.

## License

MIT
