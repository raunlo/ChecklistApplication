# CLAUDE.md

Quick reference for ChecklistApplication development. See linked docs for details.

## Project Overview

ChecklistApplication is a Go REST API for managing checklists with real-time updates via Server-Sent Events (SSE). Uses PostgreSQL, Google SSO, and Clean Architecture with Wire dependency injection.

## Quick Links

- **[SETUP.md](SETUP.md)** - Dev commands, build, run locally, database
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Layer structure, design patterns, code generation
- **[API.md](API.md)** - Frontend integration, endpoints, SSE, auth
- **[HANDOFF.md](HANDOFF.md)** - Current work status, templates feature progress

## Essential Commands

```bash
./generate.sh              # After modifying openapi/api_v1.yaml or wire.go
go test ./...             # Run all tests
go build ./...            # Build all packages
docker-compose up         # Run locally (backend + PostgreSQL)
```

## Key Architectural Decisions

| Concept | Pattern | Why |
|---------|---------|-----|
| **Layers** | server → service → repository | Clean Architecture, testability |
| **Auth** | Cookie-based Google SSO | Stateless, no session storage |
| **Real-time** | In-memory SSE broker | Simple, no external deps |
| **Ordering** | Doubly-linked list | Fast reordering without renumbering |
| **Access Control** | Guard rails (404 for denied) | Security: don't reveal resource existence |

## Layer Structure (Quick)

```
internal/
├── server/        # HTTP handlers (thin layer)
├── core/
│   ├── service/   # Business logic
│   ├── domain/    # Entities (no deps)
│   └── repository/ # Interfaces
└── repository/    # PostgreSQL impl
```

**Rule**: Imports only interfaces, never concrete types across layers.

## Never Do These

- Edit `*_gen.go` or `wire_gen.go` manually
- Import `internal/repository` from `internal/core/`
- Skip `./generate.sh` after API changes
- Store context in structs (pass as parameter)
- Return 403 for access denied (use 404 instead)

## Important Notes

- Always run `./generate.sh` after modifying OpenAPI or Wire
- Context must contain User ID (via JWT middleware)
- Client ID header must be unique per client instance (SSE echo prevention)
- Guard rails return 404 for security (don't reveal existence)

## File Locations Quick Ref

- Controllers: `internal/server/v1/`
- Services: `internal/core/service/`
- Domain entities: `internal/core/domain/`
- Repositories: `internal/repository/`
- SQL queries: `internal/repository/query/`
- OpenAPI spec: `openapi/api_v1.yaml`
- Wire config: `internal/deployment/wire.go`

---

Always update CLAUDE.md when workflow or essential patterns change. For large changes, create separate docs and link from here.
