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
- MongoDB for append-only audit events
- slog JSON logs
- `godotenv` for local configuration

## Layout

```
cmd/api/            HTTP server entrypoint
database/           connection pool
internal/config/    environment configuration
internal/dto/       request types
internal/handler/   HTTP adapters
internal/obs/       slog and request-id middleware
internal/audit/     async Mongo event writer
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
| `PORT`             | HTTP listen port                                 | `8081`   |
| `DSN`              | PostgreSQL connection string (Compose publishes Postgres on host port **5433**) | required |
| `MONGO_URI`        | MongoDB connection string (Compose publishes Mongo on host port **27018**) | optional |
| `MONGO_DB`         | Audit database name                              | `urlshortener` |
| `MONGO_COLLECTION` | Audit collection name                            | `events` |

The DSN must point at the same database the write service uses.

## Run locally

From the repository root, start PostgreSQL and MongoDB:

```bash
docker compose up -d postgres mongo
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

If `MONGO_URI` is unset or Mongo is unreachable, redirects still succeed. Audit events are dropped until Mongo is available.

## Observability

Each request gets an `X-Request-ID`. Access logs are JSON on stdout. Resolve attempts are appended asynchronously to Mongo (`urlshortener.events`) and must not delay the 302.

## Tests

Read-service tests live in `tests/`:

```bash
go test ./tests
```
