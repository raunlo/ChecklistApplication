# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ChecklistApplication is a Go-based REST API service for managing checklists with real-time updates via Server-Sent Events (SSE). It uses PostgreSQL for persistence, Google SSO for authentication, and follows Clean Architecture principles with dependency injection via Wire.

## Build and Development Commands

### Code Generation
```bash
# Run both OpenAPI and Wire code generation (required after API or dependency changes)
./generate.sh

# Or manually:
go generate -tags="oapigen" ./...    # Generate OpenAPI server code first
go generate -tags="wireinject" ./... # Generate Wire dependency injection code
```

### Build and Test
```bash
# Build
go build ./...

# Run tests
go test ./...

# Run specific test
go test ./internal/core/service -v -run TestSpecificFunction

# Vet
go vet ./...

# Format
gofmt -w .
```

### Running the Application
```bash
# Local (requires PostgreSQL running)
go run cmd/app.go

# Docker Compose (recommended)
docker-compose up

# Build Docker image
docker build -t checklistapp .
```

### Database
- PostgreSQL 14.2+
- Schema defined in `init.sql`
- Uses doubly-linked list structure for checklist item ordering (NEXT_ITEM_ID/PREV_ITEM_ID)
- Recursive CTE view `CHECKLIST_ITEMS_ORDERED_VIEW` for retrieving ordered items

## Architecture

### Layer Structure

```
cmd/                        # Application entry point
internal/
├── deployment/             # Dependency injection and app bootstrapping
│   ├── wire.go            # Wire DI configuration (generates wire_gen.go)
│   ├── gin_configuration.go # Gin router setup with CORS and auth
│   └── app.go             # Application struct and startup
├── core/                   # Business logic (framework-independent)
│   ├── domain/            # Domain entities and value objects
│   ├── service/           # Business logic services
│   ├── repository/        # Repository interfaces
│   ├── guard_rail/        # Authorization guards (ownership checks)
│   └── notification/      # SSE broker and notification service
├── repository/             # Repository implementations (PostgreSQL)
│   ├── query/             # SQL queries
│   └── connection/        # Database connection management
└── server/                 # HTTP layer
    ├── auth/              # JWT validation (Google SSO)
    └── v1/                # API v1 handlers
        ├── checklist/     # Checklist endpoints
        ├── checklistItem/ # Checklist item endpoints
        └── sse/           # SSE endpoint for real-time updates
openapi/
└── api_v1.yaml            # OpenAPI 3.0 specification
```

### Key Architectural Patterns

**Clean Architecture with Dependency Injection:**
- `internal/core` contains framework-agnostic business logic
- Dependencies flow inward: server → service → repository
- Wire generates dependency injection code from `internal/deployment/wire.go`

**Code Generation:**
- OpenAPI specs (`openapi/api_v1.yaml`) generate server interfaces via `oapi-codegen`
- Controllers implement generated interfaces (e.g., `ServerInterface` in `server.gen.go`)
- Each endpoint group (checklist, checklistItem, sse) has its own config in `cfg.yaml`

**Authentication & Authorization:**
- Cookie-based auth with Google SSO JWT validation (`internal/server/auth/jwt.go`)
- User ID extracted from JWT and stored in context via middleware
- Guard rails (`internal/core/guard_rail`) check ownership before operations
- Client ID header (`X-Client-Id`) tracks which client made changes

**Real-time Updates:**
- In-memory SSE broker (`internal/core/notification/notification_service.go`)
- Services publish events after mutations (create/update/delete/reorder)
- Broker filters events by Client ID to prevent echo to originating client
- Non-blocking publish with buffered channels (drops events if client slow)

**Database Ordering:**
- Checklist items use doubly-linked list (NEXT_ITEM_ID/PREV_ITEM_ID columns)
- Reordering updates pointer fields, not order numbers
- `CHECKLIST_ITEMS_ORDERED_VIEW` reconstructs order via recursive CTE

## Configuration

Application configured via `application.yaml` with environment variable substitution:
- Database: `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `DATABASE_SCHEMA`
- Server: port defaults to 8080
- CORS: `CORS_CONFIGURATION_HOST_NAME` (defaults to `*`, should be restricted in production)
- Google SSO: `GOOGLE_SSO_CLIENT_ID`

Format: `${VAR_NAME:default_value}`

## Frontend Integration

The backend exposes a RESTful API with real-time updates via Server-Sent Events. Frontend applications (e.g., the Next.js `studio` app) integrate as follows:

### Authentication

**Cookie-Based Google SSO:**
- Frontend must obtain a Google ID token and set it as an HttpOnly cookie named `user_token`
- Cookie must have: `HttpOnly=true`, `Secure=true` (in production), `SameSite=Lax/Strict`
- Backend validates JWT on every request via `GoogleAuthMiddleware` (internal/server/auth/jwt.go:102)
- OPTIONS requests bypass auth for CORS preflight
- HTTPS enforced in production (checks `X-Forwarded-Proto` header)

**Rate Limiting:**
- 100 requests per IP per time window (configurable in middleware)
- Returns 429 Too Many Requests when exceeded

### CORS Configuration

CORS allows credentials (cookies) with these settings (internal/deployment/gin_configuration.go):
- **Allowed Origins:** Configured via `CORS_CONFIGURATION_HOST_NAME` env var
- **Development Origins:** `http://localhost:3000`, `http://localhost:9002`, `http://app.dailychexly.local.com:9002` (only in non-release mode)
- **Allowed Headers:** `Origin`, `Content-Type`, `Authorization`, `X-Client-Id`, `Cookie`
- **Allow Credentials:** `true` (required for cookie auth and SSE)
- **Allowed Methods:** GET, POST, PUT, PATCH, DELETE, OPTIONS

Frontend must send requests with `credentials: 'include'` (fetch) or `withCredentials: true` (axios).

### X-Client-Id Header

**Critical for SSE echo prevention:**
- Frontend MUST send a unique `X-Client-Id` header with every mutating request
- Backend extracts Client ID from header or query param (internal/server/server_utils/context.go:19)
- SSE broker uses Client ID to filter events: does NOT send events back to the originating client
- Recommended: Generate UUID on app load and persist per browser tab/window
- Required for: POST, PUT, PATCH, DELETE operations
- For SSE connections, can be sent as query param: `/v1/events/checklist-item-updates/{checklistId}?clientId=...`

### API Endpoints

**Base URL:** `http://localhost:8080` (development) or production domain

**Checklists:**
- `GET /api/v1/checklists` - Get all checklists for authenticated user
- `POST /api/v1/checklists` - Create a new checklist
- `GET /api/v1/checklists/{checklistId}` - Get checklist by ID
- `PUT /api/v1/checklists/{checklistId}` - Update checklist
- `DELETE /api/v1/checklists/{checklistId}` - Delete checklist

**Checklist Items:**
- `GET /api/v1/checklists/{checklistId}/items` - Get all items (query params: `sort=asc|desc`, `completed=true|false`)
- `POST /api/v1/checklists/{checklistId}/items` - Create item
- `GET /api/v1/checklists/{checklistId}/items/{itemId}` - Get item by ID
- `PUT /api/v1/checklists/{checklistId}/items/{itemId}` - Update item
- `DELETE /api/v1/checklists/{checklistId}/items/{itemId}` - Delete item
- `POST /api/v1/checklists/{checklistId}/items/{itemId}/change-order` - Reorder item
- `POST /api/v1/checklists/{checklistId}/items/{itemId}/rows` - Add row to item
- `DELETE /api/v1/checklists/{checklistId}/items/{itemId}/rows/{rowId}` - Delete row

All endpoints require `X-Client-Id` header (defined as optional in OpenAPI but required for SSE to work correctly).

### Server-Sent Events (SSE)

**Endpoint:** `GET /v1/events/checklist-item-updates/{checklistId}`

**Parameters:**
- Path: `checklistId` (uint, required)
- Query: `clientId` (string, optional) - alternative to `X-Client-Id` header

**Connection Setup:**
```javascript
const eventSource = new EventSource(
  `http://localhost:8080/v1/events/checklist-item-updates/${checklistId}?clientId=${clientId}`,
  { withCredentials: true } // Required for cookie auth
);
```

**Event Format:**
All events are sent as JSON in the SSE `data:` field with structure:
```typescript
{
  type: string;  // Event type identifier
  payload: any;  // Type-specific payload
}
```

**Event Types and Payloads:**

1. **checklistItemCreated**
   - Payload: `ChecklistItemResponse` (full item object)

2. **checklistItemUpdated**
   - Payload: `ChecklistItemResponse` (full item object)

3. **checklistItemDeleted**
   - Payload: `{ itemId: number }`

4. **checklistItemReordered**
   - Payload: `{ itemId: number, newOrderNumber: number, orderChanged: boolean }`

5. **checklistItemRowAdded**
   - Payload: `{ itemId: number, row: ChecklistItemRowResponse }`

6. **checklistItemRowDeleted**
   - Payload: `{ itemId: number, rowId: number }`

**Event Schema Reference:** See `EventEnvelope` in `openapi/api_v1.yaml:948`

**Connection Behavior:**
- Backend sends `:ok\n\n` comment on successful connection
- Sends `:error\n\n` on JSON marshal errors
- Automatically disconnects when client closes connection or auth fails
- Guard rail checks: verifies user has access to checklist on subscribe
- Channel buffer: 10 events (events dropped if client slow, check backend logs for "dropping event")

**Frontend SSE Hook Pattern (from REVIEW.md):**
```typescript
// Simplified example based on studio/src/hooks/use-sse.ts
const useSSE = (checklistId: number, clientId: string) => {
  useEffect(() => {
    const eventSource = new EventSource(
      `/v1/events/checklist-item-updates/${checklistId}?clientId=${clientId}`,
      { withCredentials: true }
    );

    eventSource.onmessage = (event) => {
      const { type, payload } = JSON.parse(event.data);
      // Handle event types...
    };

    return () => eventSource.close();
  }, [checklistId, clientId]);
};
```

### Error Handling

**Standard Error Response:**
```json
{
  "error": "Error title",
  "message": "Detailed error message"
}
```

**Status Codes:**
- `200` - Success
- `201` - Created
- `400` - Bad request (validation error)
- `401` - Unauthorized (missing/invalid JWT)
- `403` - Forbidden (HTTPS required in production)
- `404` - Not found (or user lacks access due to guard rails)
- `429` - Too many requests (rate limit)
- `500` - Internal server error

**Guard Rail Behavior:**
- If user lacks access to a resource, backend returns `404 Not Found` (not 403) for security
- Check logs for "GuardRail: User access" messages when debugging authorization

### Development Setup

**Running Backend:**
```bash
docker-compose up  # Starts API on :8080 and PostgreSQL
```

**Frontend Environment Variables:**
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
GOOGLE_CLIENT_ID=your-google-client-id
```

**Local Development CORS:**
Backend automatically allows `localhost:3000` and `localhost:9002` in development mode.

## Testing Strategy

- Unit tests for services with mocked repositories (e.g., `checklist_items_service_test.go`)
- Table-driven tests for complex logic (e.g., `change_order_test.go` for reordering)
- Controller tests for HTTP handlers (e.g., `checklist_item_controller_impl_test.go`)
- No integration tests in CI currently

## Common Workflows

### Adding a New Endpoint

1. Update `openapi/api_v1.yaml` with new operation
2. Run `./generate.sh` to regenerate server interfaces
3. Implement interface method in controller (e.g., `internal/server/v1/checklist/checklist_controller_impl.go`)
4. Add business logic to service layer (`internal/core/service`)
5. Add repository method if needed (`internal/repository` + `internal/core/repository` interface)
6. Wire dependencies in `internal/deployment/wire.go` if adding new types

### Modifying Domain Logic

1. Update domain entity in `internal/core/domain`
2. Update service in `internal/core/service`
3. Update repository interface in `internal/core/repository` and implementation in `internal/repository`
4. Add/update tests in `*_test.go` files
5. If mutation triggers SSE event, call `notificationService.Notify*` in service

### Debugging SSE Issues

- Check Client ID is sent in `X-Client-Id` header
- Verify broker doesn't echo events to originating client (see `Publish` logic)
- Check channel buffer size (10 events) - increases if drops are logged
- SSE connections require `HasAccessToChecklist` guard rail check on subscribe

### Database Schema Changes

1. Update `init.sql` with ALTER/CREATE statements
2. Consider migration strategy for existing deployments
3. Update DBO structs in `internal/repository/dbo`
4. Update queries in `internal/repository/query`
5. Update domain entities in `internal/core/domain` if model changes

## Important Notes

- Always run `./generate.sh` after modifying `openapi/api_v1.yaml` or `internal/deployment/wire.go`
- Never edit `*_gen.go` or `wire_gen.go` files manually (they are generated)
- Context must contain User ID (via JWT middleware) for guard rails to work
- Client ID header must be unique per client instance to prevent SSE echo
- Linked list ordering: reorder operations can fail if concurrent modifications occur (no locking currently)
