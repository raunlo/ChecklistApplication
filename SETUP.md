# Setup & Development Commands

## Code Generation

After modifying `openapi/api_v1.yaml` or `internal/deployment/wire.go`:

```bash
./generate.sh
```

Or manually:
```bash
go generate -tags="oapigen" ./...    # OpenAPI server code first
go generate -tags="wireinject" ./... # Wire dependency injection code
```

## Build and Test

```bash
go build ./...
go test ./...
go test ./internal/core/service -v -run TestSpecificFunction
go vet ./...
gofmt -w .
```

## Running the Application

**Local** (requires PostgreSQL):
```bash
go run cmd/app.go
```

**Docker Compose** (recommended):
```bash
docker-compose up
```

**Build Docker image:**
```bash
docker build -t checklistapp .
```

## Database

- PostgreSQL 14.2+
- Schema: `init.sql`
- Uses doubly-linked list (NEXT_ITEM_ID/PREV_ITEM_ID) for item ordering
- Recursive CTE view: `CHECKLIST_ITEMS_ORDERED_VIEW` retrieves ordered items

## Configuration

**Via `application.yaml` with env var substitution** (`${VAR_NAME:default_value}`):
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `DATABASE_SCHEMA`
- `CORS_CONFIGURATION_HOST_NAME` (defaults to `*`)
- `GOOGLE_SSO_CLIENT_ID`
- Server port defaults to 8080

## Frontend Development

**Backend running:**
```bash
docker-compose up
```

**Frontend environment** (e.g., `studio/.env.local`):
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
GOOGLE_CLIENT_ID=your-google-client-id
```

**CORS**: Backend allows `localhost:3000` and `localhost:9002` in development.

## Testing Strategy

- **Unit tests** for services with mocked repositories
- **Table-driven tests** for complex logic
- **Controller tests** for HTTP handlers
- No integration tests in CI
