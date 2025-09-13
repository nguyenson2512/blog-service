# Blog Service

High-performance blog API in Go using Gin, GORM (PostgreSQL), Redis (cache-aside), and Elasticsearch (full-text search). Hot reloaded with Air. Orchestrated via Docker Compose.

## Stack
- Go + Gin (HTTP)
- GORM + PostgreSQL (data + migrations)
- Redis (cache-aside pattern, TTL 5m)
- Elasticsearch (full-text search on title and content)
- Air (hot reload in container)

## Quickstart
1. Copy the example env and adjust if needed:
```bash
cp .env.example .env
```
2. Build and start everything:
```bash
docker compose up --build
```
3. Wait for services to become healthy. The API will be available at:
- Base URL: `http://localhost:8080`

To stop:
```bash
docker compose down
```
To stop and remove data volumes:
```bash
docker compose down -v
```

## Environment
The app reads environment variables from `.env` (see `.env.example`). Key values:
- PostgreSQL: host `postgres`, user `postgres`, password `postgres`, db `blog`
- Redis: `redis:6379`
- Elasticsearch: `http://elasticsearch:9200`

## Database
- Tables: `posts`, `activity_logs`
- `posts.tags` is a `TEXT[]`. A GIN index is created on boot to optimize tag search:
  - `CREATE INDEX IF NOT EXISTS idx_posts_tags_gin ON posts USING GIN (tags);`
- Migrations are handled by GORM AutoMigrate at startup.

## Caching (Cache-Aside)
- GET `/posts/:id` first checks Redis (`post:<id>`). TTL is 300 seconds.
- PUT `/posts/:id` invalidates the Redis key to ensure subsequent reads hit the database before being re-cached.

## Elasticsearch
- Index: `posts`
- On create/update, the document `{id,title,content,tags}` is indexed (refresh=true).
- Search uses `multi_match` across `title` and `content`.
- Related posts use `bool` query with `should` clauses for tag matching.

## API Reference and Sample Requests
Use `jq` for pretty-printing where shown.

### Create a post (transactional with activity log + ES index)
POST `/posts`
```bash
curl -sS -X POST http://localhost:8080/posts \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Hello World",
    "content": "This is my first post.",
    "tags": ["golang", "news"]
  }' | jq
```
Response (201):
```json
{
  "id": 1,
  "title": "Hello World",
  "content": "This is my first post.",
  "tags": ["golang", "news"],
  "created_at": "...",
  "updated_at": "..."
}
```

### Get a post by ID (Redis cache-aside)
GET `/posts/:id`
```bash
curl -sS http://localhost:8080/posts/1 | jq
```

### Get a post with related posts (based on shared tags)
GET `/posts/:id?include_related=true`
```bash
curl -sS 'http://localhost:8080/posts/1?include_related=true' | jq
```
Response includes up to 5 related posts that share similar tags:
```json
{
  "id": 1,
  "title": "Hello World",
  "content": "This is my first post.",
  "tags": ["golang", "news"],
  "created_at": "...",
  "updated_at": "...",
  "related_posts": [
    {
      "id": 3,
      "title": "Go Performance Tips",
      "content": "Some performance tips...",
      "tags": ["golang", "performance"]
    },
    {
      "id": 5,
      "title": "Latest Tech News",
      "content": "Breaking news...",
      "tags": ["news", "tech"]
    }
  ]
}
```

### Update a post (invalidates Redis, re-indexes ES)
PUT `/posts/:id`
```bash
curl -sS -X PUT http://localhost:8080/posts/1 \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Hello World (Edited)",
    "content": "Updated content.",
    "tags": ["golang", "update"]
  }' | jq
```

### Search posts by tag (uses GIN index on TEXT[])
GET `/posts/search-by-tag?tag=<tag>`
```bash
curl -sS 'http://localhost:8080/posts/search-by-tag?tag=golang' | jq
```

### Full-text search (ES multi_match: title, content)
GET `/posts/search?q=<query>`
```bash
curl -sS 'http://localhost:8080/posts/search?q=hello' | jq
```

## Related Posts Feature
The related posts feature uses Elasticsearch to find posts with similar tags:

- **Algorithm**: Uses a `bool` query with `should` clauses for each tag
- **Exclusion**: Excludes the current post using `must_not` clause
- **Limit**: Returns up to 5 related posts
- **Fallback**: If no related posts found or ES error, returns empty array
- **Usage**: Add `?include_related=true` to GET `/posts/:id`

The Elasticsearch query structure:
```json
{
  "query": {
    "bool": {
      "should": [
        {"term": {"tags": "golang"}},
        {"term": {"tags": "news"}}
      ],
      "must_not": [
        {"term": {"id": 1}}
      ],
      "minimum_should_match": 1
    }
  },
  "size": 5
}
```

## Development Workflow
- The API container runs `air` for hot reloading.
- Source is mounted into the container; edits trigger rebuilds automatically.
- Logs are visible in the `docker compose up` output.

## Health Checks (compose)
- Postgres: `pg_isready`
- Redis: `redis-cli ping`
- Elasticsearch: `GET /` (no auth)

## Troubleshooting
- Ports in use: ensure `5432`, `6379`, `9200`, `8080` are free on your host.
- Elasticsearch memory: increase Docker resources if ES fails to start.
- Reset state: `docker compose down -v` to remove volumes and start fresh.
- Dependency download issues: run `docker compose build --no-cache`.

## Project Layout (key paths)
- `cmd/app/main.go` — entrypoint
- `internal/app/app.go` — wiring (config, DB, Redis, ES, router)
- `internal/config` — env config
- `internal/db` — GORM setup, migrations, GIN index
- `internal/models` — `Post`, `ActivityLog`
- `internal/repository` — data access
- `internal/service` — business logic (transactions, cache-aside, ES sync)
- `internal/search` — Elasticsearch client wrapper
- `internal/transport/http` — router and HTTP layer
- `internal/transport/http/handlers` — Gin handlers

## Cleaning Up
```bash
docker compose down -v
```
This stops containers and removes volumes (Postgres data will be deleted).
