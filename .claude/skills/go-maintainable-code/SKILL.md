---
name: go-maintainable-code
description: Write clean, maintainable Go code following Clean Architecture, dependency injection, and ChecklistApplication patterns. Use when writing new Go code, refactoring, or implementing features.
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
---

# Go Maintainable Code Skill

Write Go code following ChecklistApplication patterns: Clean Architecture, interface-based design, and dependency injection.

## Core Principles

**See [ARCHITECTURE.md](../../../ARCHITECTURE.md) for detailed patterns.**

Quick reference:
- Dependency flow: `server → service → repository`
- Import interfaces only, never concrete types across layers
- Service/domain have no external dependencies
- Use Wire for dependency injection

## Error Handling

```go
// Use domain.Error (custom error type)
return domain.NewError("message", 400)
return domain.Wrap(err, "context", 500)

// Guard rails return 404 for access denied
if err := guardrail.HasAccess(ctx, id); err != nil {
    return error.NewNotFoundError(id)  // 404, not 403
}
```

## Testing Pattern

```go
func TestService_Method_Scenario(t *testing.T) {
    mockRepo := new(mockRepository)
    mockRepo.On("Query", mock.Anything, args).Return(expected, nil)

    svc := &service{repository: mockRepo}
    result, err := svc.Method(context.Background(), args)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    mockRepo.AssertExpectations(t)
}
```

## Struct Pattern

```go
// Public interface
type IMyService interface {
    DoWork(ctx context.Context) error
}

// Private implementation
type myService struct {
    repo repository.IMyRepository
}

// Public constructor (for Wire)
func NewMyService(repo repository.IMyRepository) IMyService {
    return &myService{repo: repo}
}
```

## Code Generation Workflow

1. Update `openapi/api_v1.yaml`
2. Run `./generate.sh`
3. Implement generated interface in controller
4. **NEVER** edit `*_gen.go` files

## Checklist Before Committing

- [ ] Follows Clean Architecture layers
- [ ] Uses interfaces for dependencies
- [ ] Has unit tests with mocks
- [ ] Guard rails check authorization
- [ ] Ran `./generate.sh` if API changed
- [ ] `go test ./...` passes
- [ ] `go build ./...` succeeds
- [ ] Meaningful error messages

## Related Documentation

- [ARCHITECTURE.md](../../../ARCHITECTURE.md) - Full layer structure & patterns
- [SETUP.md](../../../SETUP.md) - Dev commands
- [HANDOFF.md](../../../HANDOFF.md) - Current work in progress
