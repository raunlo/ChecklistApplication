# Work Handoff: Templates Feature Implementation

## Current Status

**Phase**: 2 (Backend HTTP Controllers) - 60% complete
**Branch**: `templates-uus`
**Last Updated**: 2026-03-23

## Completed ✅

**Database & Domain** (Phase 1):
- [x] Database schema: `TEMPLATE` and `TEMPLATE_ITEM` tables with sequences and indexes
- [x] Domain entities: `internal/core/domain/template.go`
- [x] Repository interface: `internal/core/repository/template_repository.go`
- [x] Repository implementation: `internal/repository/template_repository.go`
- [x] SQL queries: `internal/repository/query/template_queries.go`
- [x] Data objects: `internal/repository/dbo/template_dbo.go`

**Service Layer** (Phase 1):
- [x] Service implementation: `internal/core/service/template_service.go`
- [x] Guard rails: `internal/core/guard_rail/template_ownership_checker.go`
- [x] Key methods:
  - `SaveTemplate` - Create or update template
  - `FindTemplateById` - Retrieve with validation
  - `ApplyTemplateToChecklist` - Add items to existing checklist
  - `CreateChecklistFromTemplate` - Create new checklist from template
  - `GetTemplatePreview` - Show items without applying
- [x] Unit tests pass
- [x] Code compiles

**OpenAPI & Code Generation**:
- [x] OpenAPI spec: `openapi/api_v1.yaml` (7 endpoints, 5 schemas)
- [x] Spec is complete and valid

## In Progress ⏳

**Phase 2: HTTP Controllers**:
- [ ] Generate server interfaces: `./generate.sh`
- [ ] Implement template controller at: `internal/server/v1/template/`
- [ ] Create DTO mapper: Convert domain models ↔ API responses
- [ ] Wire dependencies in `internal/deployment/wire.go`

## Next Steps After Phase 2

**Phase 3: Frontend Integration**:
- [ ] Add API calls in Next.js `studio` app
- [ ] Create template UI pages
- [ ] Implement SSE for template real-time updates (if needed)
- [ ] Test end-to-end

## Key Files & Locations

```
internal/
├── core/
│   ├── domain/template.go                          # Domain entities
│   ├── repository/template_repository.go           # Interface
│   ├── service/template_service.go                 # Business logic
│   └── guard_rail/template_ownership_checker.go   # Authorization
├── repository/
│   ├── template_repository.go                      # Implementation
│   ├── dbo/template_dbo.go                        # Data objects
│   ├── query/template_queries.go                  # SQL queries
│   └── template_repository.go
├── server/
│   └── v1/template/                               # [TO CREATE] Controllers
└── deployment/
    └── wire.go                                     # [TO UPDATE]
openapi/
└── api_v1.yaml                                     # [COMPLETE] API spec
```

## Architectural Notes

**Pattern**: Follows existing ChecklistApplication patterns:
- Clean Architecture layers (domain → service → repository)
- Dependency injection via Wire
- Guard rails for authorization (returns 404 for access denied)
- Case-insensitive duplicate detection in service

**Database**: Templates and items use same structure as checklists:
- Case-insensitive `LOWER(name)` for duplicate detection
- Foreign key constraints with CASCADE delete
- Sequences for ID generation

## Immediate TODOs

1. Run `./generate.sh` to regenerate OpenAPI server interfaces
2. Create `internal/server/v1/template/template_controller_impl.go`
3. Implement required interface methods:
   - `GetTemplates` - List all templates
   - `CreateTemplate` - Create new
   - `GetTemplate` - Get by ID
   - `UpdateTemplate` - Update
   - `DeleteTemplate` - Delete
   - `ApplyTemplate` - Apply to existing checklist
   - `PreviewTemplate` - Show items without applying
4. Create DTO mapper for Template → API response
5. Add to Wire configuration
6. Run `go test ./...` to verify
7. Commit with handoff for Phase 3

## Potential Issues to Watch

- SSE notifications: Should templates trigger real-time updates?
- Concurrent template applications: Need guard rail on write
- Item limit: Should templates have max items? (Currently no limit)
- Deletion safety: Deleting template doesn't affect created checklists (correct)

## Questions for Product/Design

- Should applying a template trigger SSE events?
- Should there be template versioning?
- Should users be able to share templates?
- Template organization: Tags, categories, folders?

## Related Documentation

- See [ARCHITECTURE.md](ARCHITECTURE.md) for layer patterns
- See [API.md](API.md) for endpoint conventions
- See [SETUP.md](SETUP.md) for dev commands
