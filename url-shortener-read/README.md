# URL Shortener Read Service

Query-side HTTP API that resolves a short code to its original URL and redirects the client.

## Responsibilities

- Look up `short_code` in PostgreSQL
- Issue an HTTP redirect to `original_url`
- Reject invalid codes and return 404 when no mapping exists

This service does not create or modify mappings. Schema is owned by the write service.

## Stack

- Go 1.25
- Gin
- PostgreSQL via `database/sql` and [pgx](https://github.com/jackc/pgx)
- `godotenv` for local configuration

## Layout

```
cmd/api/            HTTP server entrypoint
database/           connection pool
internal/config/    environment configuration
internal/dto/       request types
internal/handler/   HTTP adapters
internal/model/     domain model
internal/repository persistence (read-only)
internal/service/   code validation and lookup
routes/             route registration
tests/              read-service tests
```

Requests flow `route → handler → service → repository → PostgreSQL`.

## Configuration

Copy `.env.example` to `.env` and adjust as needed:

| Variable | Description                                      | Default  |
|----------|--------------------------------------------------|----------|
| `PORT`   | HTTP listen port                                 | `8081`   |
| `DSN`    | PostgreSQL connection string (Compose publishes Postgres on host port **5433**) | required |

The DSN must point at the same database the write service uses.

## Run locally

From the repository root, start PostgreSQL:

```bash
docker compose up -d postgres
```

Start the write service first so the schema exists, then from `url-shortener-read`:

```bash
cp .env.example .env
go run ./cmd/api
```

## API

### `GET /health`

Checks process liveness and PostgreSQL connectivity.

```json
{"service":"read","status":"ok"}
```

### `GET /:code`

Resolve a short code and redirect.

**Success** — `302 Found` with `Location` set to the original URL.

**Errors**

| Status | When |
|--------|------|
| `400`  | Short code is not 4–10 alphanumeric characters |
| `404`  | No mapping exists for the code |
| `500`  | Unexpected persistence failure |
| `503`  | Health check cannot reach PostgreSQL |

### Example

```bash
curl -sS -D - -o /dev/null http://localhost:8081/docs42
```

Typical headers:

```
HTTP/1.1 302 Found
Location: https://example.com/docs
```

## Tests

Read-service tests live in `tests/`:

```bash
go test ./tests
```
