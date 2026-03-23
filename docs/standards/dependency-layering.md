---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-03-23
source_of_truth:
  - docs/ai/repo-rules.md
  - server/internal
  - static/src
---

# Dependency Layering Standard

## Scope

This standard governs import direction between layers in both backend and frontend. It applies to all code changes in the repository.

## Backend Dependency Direction

```
model → repository → service → handler → router/middleware
  ↑                                            ↑
  pkg/* (shared infrastructure)                bootstrap/
```

### Rules

| Layer | May Import | Must Not Import |
| --- | --- | --- |
| `model` | standard library, GORM tags | `repository`, `service`, `handler`, `router`, `middleware` |
| `repository` | `model`, standard library, GORM, `pkg/*` | `service`, `handler`, `router`, `middleware` |
| `service` | `model`, `repository`, `pkg/*`, other services | `handler`, `router`, `middleware` |
| `handler` | `service`, `model` (for request/response types), `pkg/response` | `repository` directly |
| `router` | `handler`, `middleware`, `service` (for DI) | `repository` directly |
| `middleware` | `model` (for role constants), `pkg/*`, `service` (for auth) | `handler`, `repository` directly |
| `pkg/*` | standard library, external packages | `internal/*` |

### Common Violations

**Handler importing repository:**

```go
// WRONG — handler reaching past service into repository
func (h *FleetHandler) List(c *gin.Context) {
    fleets, err := h.repo.ListFleets(params)  // violation
}

// CORRECT — handler calls service
func (h *FleetHandler) List(c *gin.Context) {
    fleets, err := h.service.ListFleets(params)
}
```

**Repository containing business logic:**

```go
// WRONG — authorization decision in repository
func (r *UserRepo) GetUser(id uint, requesterRole string) (*model.User, error) {
    if requesterRole != "admin" {  // violation: business logic
        return nil, errors.New("forbidden")
    }
}

// CORRECT — repository does data access, service does authorization
func (r *UserRepo) GetUser(id uint) (*model.User, error) {
    // pure data access
}
```

**Model importing service:**

```go
// WRONG — model depending on service layer
import "amiya-eden/internal/service"

// CORRECT — model has zero internal dependencies
```

## Frontend Dependency Direction

```
types → api → hooks/store → components → views
```

### Rules

| Layer | May Import | Must Not Import |
| --- | --- | --- |
| `types/` | nothing (pure type definitions) | `api/`, `hooks/`, `store/`, `components/`, `views/` |
| `api/` | `types/`, HTTP client utilities | `hooks/`, `store/`, `components/`, `views/` |
| `hooks/` | `types/`, `api/`, `store/`, other hooks | `views/`, `components/` (specific ones) |
| `store/` | `types/`, `api/`, `hooks/` | `views/`, `components/` |
| `components/` | `types/`, `hooks/`, `store/`, other components | `views/`, `api/` directly (should go through hooks) |
| `views/` | all layers above | should not be imported by others |

### Common Violations

**View calling fetch directly:**

```typescript
// WRONG — view making HTTP calls
const response = await axios.get('/api/v1/fleets')

// CORRECT — view calls API wrapper
import { getFleets } from '@/api/fleet'
const response = await getFleets(params)
```

**API layer importing from views:**

```typescript
// WRONG — circular dependency
import { FleetFormData } from '@/views/operation/fleets/types'

// CORRECT — shared types live in types/
import type { Api } from '@/types/api/api'
```

## Cross-Boundary Rules

### Backend ↔ Frontend Contract

When a backend response shape changes:

1. Backend: update handler response / request shape
2. Backend: update service if needed
3. Frontend: update `static/src/api/*.ts`
4. Frontend: update `static/src/types/api/api.d.ts`
5. Frontend: update consuming views/components

Field names in backend JSON tags must match frontend type definitions exactly. Do not silently rename across the boundary.

### Infrastructure Layer (`pkg/*`)

`pkg/*` provides shared utilities (JWT, ESI client, response helpers) to `internal/*`. It must never import from `internal/*`. If `pkg` code needs internal types, either:

- Define the types in `pkg` and have `internal` use them
- Use interfaces for dependency injection (as done with ESI queue)

## Enforcement

### Current Mechanism

Enforcement is through code review, agent verification, and the pre-completion protocol (see "Verification and Completion" in `docs/ai/repo-rules.md` and `docs/standards/pre-completion-checklist.md`).

### What To Do When You Find a Violation

1. Flag the violation in your change summary
2. If the violation is in code you're modifying, fix it as part of your change
3. If the violation is in unrelated code, note it but do not fix it (scoped changes rule)
4. Do not introduce new violations under any circumstances

## Pre-Submission Check

Before submitting a change:

- [ ] No new imports from a lower layer to a higher layer
- [ ] `handler` does not import `repository`
- [ ] `repository` does not contain business logic
- [ ] `views` do not call HTTP directly
- [ ] `types/` has no imports from other application layers
- [ ] `pkg/*` does not import `internal/*`
