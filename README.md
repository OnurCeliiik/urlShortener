# URL Shortener

A URL shortening system split into independent **write** and **read** services. The write service records mappings from a short code to an original URL. The read service is responsible for looking those mappings up and redirecting clients.

The two sides share PostgreSQL as the source of truth. Reads can later sit behind a cache without changing how mappings are created.

## Architecture

The system follows a CQRS-style split:

```
Client
  │
  ├─ POST /shorten ──► write service ──► PostgreSQL
  │
  └─ GET  /:code   ──► read service  ──► cache ──► PostgreSQL
                              │
                              └── 302 Location: original URL
```

**Write service** owns commands: validation, optional custom codes, code generation, uniqueness, and persistence. It never redirects.

**Read service** owns queries: resolve a short code to its original URL and issue the redirect. High read volume is expected, so this side is designed to scale independently of writes.

Schema lives with the write service. Both services speak to the same `urls` table (`short_code`, `original_url`, `created_at`).

## Technology

| Area            | Choice                                      |
|-----------------|---------------------------------------------|
| Language        | Go                                          |
| Write HTTP      | Gin                                         |
| Database        | PostgreSQL                                  |
| DB driver       | pgx (`database/sql` stdlib adapter)         |
| Local database  | Docker Compose                              |
| Config          | Environment variables (`.env` for local)    |

## Repository layout

```
url-shortener-write/   command API (create mappings)
url-shortener-read/    query/redirect API
docker-compose.yml     local PostgreSQL
```

Write-service details, configuration, and HTTP contract: [url-shortener-write/README.md](url-shortener-write/README.md).

## Mapping model

- Each `short_code` is unique (primary key).
- Each `original_url` is unique: shortening the same URL again returns the existing code.
- Short codes are 4–10 alphanumeric characters. Generated codes are 7 characters.
- Clients may supply a custom code or leave it empty and let the write service generate one.

## Local PostgreSQL

```bash
docker compose up -d postgres
```

Default connection:

```
postgres://shortener:shortener@localhost:5433/urlshortener?sslmode=disable
```
