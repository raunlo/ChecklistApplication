# Architecture & Patterns

## Layer Structure

```
internal/
├── server/              # HTTP layer (Gin, OpenAPI controllers)
│   └── Depends on: service (via interfaces)
├── core/
│   ├── service/        # Business logic (framework-independent)
│   │   └── Depends on: repository interfaces, domain
│   ├── domain/         # Entities, value objects (no dependencies)
│   ├── repository/     # Repository interfaces
│   └── guard_rail/     # Authorization checks
├── repository/         # PostgreSQL implementations
│   ├── query/         # SQL queries
│   └── connection/    # Database connection
└── deployment/         # Dependency injection & bootstrapping
```

**Dependency Flow**: `server → service → repository`

**Rule**: Import interfaces only, never concrete types across layers.

## Clean Architecture Principles

**Domain Layer** (`internal/core/domain/`):
- Pure entities and value objects
- No external dependencies
- Framework-agnostic

**Service Layer** (`internal/core/service/`):
- Business logic
- Calls repository interfaces (not concrete types)
- Uses guard rails for authorization
- Publishes SSE events after mutations

**Repository Layer** (`internal/repository/`):
- Data access implementation
- SQL queries in `query/` package
- Use named parameters: `pgx.NamedArgs`
- Check rows affected for deletes

**Server Layer** (`internal/server/`):
- Thin HTTP handlers
- Delegate to services
- Map errors to responses

## Dependency Injection (Wire)

Always use Wire for dependencies. Add to `internal/deployment/wire.go`:

```go
wire.Build(
    NewMyService,
    wire.Bind(new(IMyService), new(*myService)),
)
```

Then run: `./generate.sh`

Never store context in structs—pass context as first parameter to every method.

## Code Generation (OpenAPI)

1. Update `openapi/api_v1.yaml`
2. Run `./generate.sh`
3. Implement generated interface in controller
4. **NEVER** edit `*_gen.go` files manually

## Error Handling

Use `domain.Error` (custom error interface):

```go
domain.NewError(message, statusCode)
domain.Wrap(err, context, statusCode)
```

Guard rails return **404** for unauthorized access (not 403) for security.

## Struct Patterns

**Private implementation, public interface:**
```go
type IMyService interface { ... }
type myService struct { ... }
func NewMyService(...) IMyService { ... }
```

**Testing**: Mock repository, test service behavior.

## SSE Real-Time Updates

- In-memory broker: `internal/core/notification/`
- Services publish after mutations
- Broker filters by `X-Client-Id` header (prevents echo)
- Non-blocking publish with 10-event buffer
- Guard rail check on subscribe

**Event structure**:
```json
{
  "type": "checklistItemCreated",
  "payload": { ... }
}
```

## Database Patterns

**Doubly-linked list ordering**:
- Items use `NEXT_ITEM_ID`, `PREV_ITEM_ID` columns
- Query via: `CHECKLIST_ITEMS_ORDERED_VIEW` (recursive CTE)
- Phantom items: `IS_PHANTOM = true` (filtered in queries)

**Transactions**:
```go
res, err := connection.RunInTransaction(connection.TransactionProps[ResultType]{
    Query: func(tx pool.TransactionWrapper) (ResultType, error) { ... },
    Connection: r.connection,
    TxOptions: pgx.TxOptions{IsoLevel: pgx.Serializable},
})
```

## Key Architectural Decisions

| Feature | Pattern | Why |
|---------|---------|-----|
| **Auth** | Cookie-based Google SSO JWT | Stateless, no session overhead |
| **CORS** | Allow credentials with specific origins | Secure cookie auth across domains |
| **Client tracking** | X-Client-Id header | Prevent SSE echo, detect duplicates |
| **Errors** | 404 for access denied | Security (don't reveal resource existence) |
| **Ordering** | Doubly-linked list | Fast reordering without renumbering |
| **Pub/Sub** | In-memory channels | Simple, no external dependencies |

## Common Workflows

### Adding a New Endpoint

1. Update `openapi/api_v1.yaml` with new operation
2. Run `./generate.sh`
3. Implement interface in controller
4. Add service method
5. Add repository if needed
6. Update `wire.go`

### Modifying Domain Logic

1. Update entity in `internal/core/domain/`
2. Update service in `internal/core/service/`
3. Update repository interface + implementation
4. Add/update tests
5. If mutation: call `notificationService.Notify*`

### Debugging SSE Issues

- Verify `X-Client-Id` header sent
- Check client ID filters events correctly
- Check channel buffer size (10 events)
- Verify `HasAccessToChecklist` guard rail on subscribe

## Important Notes

- Run `./generate.sh` after modifying `openapi/api_v1.yaml` or `wire.go`
- Never edit `*_gen.go` or `wire_gen.go` manually
- Context must contain User ID (via JWT middleware) for guard rails
- Client ID header must be unique per client instance
