# URL Shortener Write Service

Command-side HTTP API that creates and stores short-URL mappings in PostgreSQL.

## Responsibilities

- Accept a long URL and optionally a custom short code
- Validate input
- Persist a unique mapping of `short_code` → `original_url`
- Return the public short URL for the read/redirect side to resolve later

This service does not perform redirects.

## Stack

- Go 1.25
- Gin
- PostgreSQL via `database/sql` and [pgx](https://github.com/jackc/pgx)
- `godotenv` for local configuration

## Layout

```
cmd/api/            HTTP server entrypoint
database/           connection pool and migration runner
migrations/         SQL schema
internal/config/    environment configuration
internal/dto/       request and response types
internal/handler/   HTTP adapters
internal/model/     domain model
internal/repository persistence
internal/service/   validation, code generation, mapping rules
routes/             route registration
tests/              write-service tests
```

Requests flow `route → handler → service → repository → PostgreSQL`.

## Configuration

Copy `.env.example` to `.env` and adjust as needed:

| Variable   | Description                                      | Default                     |
|------------|--------------------------------------------------|-----------------------------|
| `PORT`     | HTTP listen port                                 | `8080`                      |
| `DSN`      | PostgreSQL connection string (Compose publishes Postgres on host port **5433**) | required |
| `BASE_URL` | Public origin of the read/redirect service, used to build `short_url` values | `http://localhost:$PORT`    |

`BASE_URL` should be the read service origin (default `http://localhost:8081`).

## Run locally

From the repository root, start PostgreSQL:

```bash
docker compose up -d postgres
```

Then from `url-shortener-write`:

```bash
cp .env.example .env
go run ./cmd/api
```

The process applies schema migrations on startup.

## API

### `GET /health`

Checks process liveness and PostgreSQL connectivity.

```json
{"service":"write","status":"ok"}
```

### `POST /shorten`

Create or reuse a mapping.

**Request**

```json
{
  "original_url": "https://example.com/very/long/path",
  "short_code": "docs42"
}
```

`short_code` is optional. When omitted, the service generates a 7-character alphanumeric code.

**Success (created)** — `201 Created`

```json
{
  "short_code": "docs42",
  "short_url": "http://localhost:8081/docs42",
  "original_url": "https://example.com/very/long/path"
}
```

**Success (existing mapping reused)** — `200 OK` with the same body shape.

**Errors**

| Status | When |
|--------|------|
| `400`  | Invalid JSON, URL, or short code |
| `409`  | Short code already used by another URL, or the URL already has a different code |
| `500`  | Unexpected persistence failure |
| `503`  | Health check cannot reach PostgreSQL, or a unique code could not be allocated |

Rules:

- `original_url` must be an `http` or `https` URL, max 2048 characters
- Custom `short_code` must be 4–10 alphanumeric characters
- Each original URL maps to exactly one short code
- Repeating `POST /shorten` with the same URL returns the existing code

### Example

```bash
curl -sS -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"original_url":"https://example.com/very/long/path"}'
```

Custom code:

```bash
curl -sS -X POST http://localhost:8080/shorten \
  -H 'Content-Type: application/json' \
  -d '{"original_url":"https://example.com/docs","short_code":"docs42"}'
```

## Tests

Write-service tests live in `tests/`:

```bash
go test ./tests
```
