# Frontend Integration Guide

## Authentication

**Cookie-Based Google SSO:**
- Frontend must set `user_token` HttpOnly cookie with Google ID token
- Cookie must have: `HttpOnly=true`, `Secure=true` (production), `SameSite=Lax/Strict`
- Backend validates JWT via `GoogleAuthMiddleware` on every request
- OPTIONS requests bypass auth (CORS preflight)
- HTTPS enforced in production (checks `X-Forwarded-Proto` header)

## CORS Configuration

**Settings** (`internal/deployment/gin_configuration.go`):
- **Allowed Origins**: Configured via `CORS_CONFIGURATION_HOST_NAME` env var
- **Dev Origins**: `localhost:3000`, `localhost:9002`, `app.dailychexly.local.com:9002` (non-release only)
- **Allowed Headers**: `Origin`, `Content-Type`, `Authorization`, `X-Client-Id`, `Cookie`
- **Allow Credentials**: `true` (required for cookie auth and SSE)
- **Methods**: GET, POST, PUT, PATCH, DELETE, OPTIONS

**Frontend**: Send requests with `credentials: 'include'` (fetch) or `withCredentials: true` (axios).

## X-Client-Id Header (Critical for SSE)

**Required for all mutating requests** (POST, PUT, PATCH, DELETE):
- Must be unique per browser tab/window
- Backend uses it to filter SSE events (prevents echo)
- Recommend: Generate UUID on app load, persist per tab

**For SSE connections**, can also be sent as query param:
```
/v1/events/checklist-item-updates/{checklistId}?clientId=...
```

## API Endpoints

**Base URL**: `http://localhost:8080` (dev) or production domain

### Checklists
- `GET /api/v1/checklists` - Get all (authenticated user)
- `POST /api/v1/checklists` - Create new
- `GET /api/v1/checklists/{checklistId}` - Get by ID
- `PUT /api/v1/checklists/{checklistId}` - Update
- `DELETE /api/v1/checklists/{checklistId}` - Delete

### Checklist Items
- `GET /api/v1/checklists/{checklistId}/items` - Get all (query: `sort=asc|desc`, `completed=true|false`)
- `POST /api/v1/checklists/{checklistId}/items` - Create
- `GET /api/v1/checklists/{checklistId}/items/{itemId}` - Get by ID
- `PUT /api/v1/checklists/{checklistId}/items/{itemId}` - Update
- `DELETE /api/v1/checklists/{checklistId}/items/{itemId}` - Delete
- `POST /api/v1/checklists/{checklistId}/items/{itemId}/change-order` - Reorder
- `POST /api/v1/checklists/{checklistId}/items/{itemId}/rows` - Add row
- `DELETE /api/v1/checklists/{checklistId}/items/{itemId}/rows/{rowId}` - Delete row

All endpoints require `X-Client-Id` header.

## Server-Sent Events (SSE)

**Endpoint**: `GET /v1/events/checklist-item-updates/{checklistId}`

**Parameters**:
- Path: `checklistId` (uint, required)
- Query: `clientId` (string, optional - alternative to header)

**Connection Setup**:
```javascript
const eventSource = new EventSource(
  `http://localhost:8080/v1/events/checklist-item-updates/${checklistId}?clientId=${clientId}`,
  { withCredentials: true } // Required for cookie auth
);
```

**Event Format** (all events sent as JSON in SSE `data:` field):
```typescript
{
  type: string;   // Event type identifier
  payload: any;   // Type-specific payload
}
```

**Event Types**:

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

**Connection Behavior**:
- Backend sends `:ok\n\n` on successful connection
- Sends `:error\n\n` on JSON marshal errors
- Automatically disconnects when client closes or auth fails
- Guard rail check verifies user has access to checklist
- Channel buffer: 10 events (events dropped if client slow - check logs)

**Frontend Hook Pattern**:
```typescript
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

## Rate Limiting

- 100 requests per IP per time window
- Returns 429 Too Many Requests when exceeded

## Error Handling

**Standard Error Response**:
```json
{
  "error": "Error title",
  "message": "Detailed error message"
}
```

**Status Codes**:
- `200` - Success
- `201` - Created
- `400` - Bad request (validation error)
- `401` - Unauthorized (missing/invalid JWT)
- `403` - Forbidden (HTTPS required in production)
- `404` - Not found (or access denied via guard rails)
- `429` - Too many requests (rate limit)
- `500` - Internal server error

**Guard Rail Behavior**: If user lacks access, backend returns `404 Not Found` (not 403) for security.
