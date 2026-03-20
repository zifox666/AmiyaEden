# AGENTS.md

Status: Active  
Scope: entire repository  
Canonical copy: this file.

## 1. Project Intent

`AmiyaEden` is a full-stack EVE Online operations platform with:

- Go backend under `server/`
- Vue 3 + TypeScript frontend under `static/`
- RBAC roles, menus, and button permissions
- dynamic menu / routing support
- ESI / SSO integrations
- strongly typed frontend API contracts

The active product authentication flow is EVE SSO-based. Legacy template pages may still exist in `static/src/views/auth/`, but they are not the current supported login architecture and should not be treated as product requirements unless the user explicitly asks for them.

Changes should preserve the existing architecture. Prefer consistency with the current repo over introducing new patterns.

## 2. Architecture Rules

### 2.1 Backend Layering

Backend flow must remain:

`router -> middleware -> handler -> service -> repository -> model`

Standards:

- `handler` is transport-only.
  - parse request
  - read auth context
  - call service
  - return standardized response
- `service` owns business rules.
  - authorization decisions beyond simple route guards
  - orchestration across repositories
  - ESI / SSO / external integration
  - response shaping for frontend use
- `repository` owns database access only.
  - no business policy
  - no HTTP calls
  - no Gin types
- `model` defines persistence and JSON contracts.
  - keep naming explicit
  - keep frontend / backend field names aligned

Do not put business logic in handlers or SQL shaping directly in handlers.

### 2.2 Frontend Layering

Frontend flow must remain:

`view -> api -> backend`

Supporting layers:

- shared logic in `static/src/hooks`
- app state in `static/src/store`
- route logic in `static/src/router`
- reusable UI in `static/src/components`
- type contracts in `static/src/types`

Standards:

- views should not call `fetch` / `axios` directly
- views should not duplicate backend contract types inline
- reusable table / list logic should prefer existing abstractions such as `useTable`
- routing, auth, and permission logic belongs in router / store / directives, not page-local hacks

### 2.3 Routing and Menu Modes

The frontend currently supports both:

- `frontend` mode via route modules under `static/src/router/modules`
- `backend` mode via `/api/v1/menu/list`

Changes to roles, menus, and button permissions must keep these aligned:

- backend route protection
- menu seeds in `server/internal/model/menu.go`
- frontend route metadata
- button permission usage such as `v-auth`

## 3. API Contract Standards

The frontend and backend are tightly coupled. Keep contracts synchronized.

When changing an endpoint:

1. update backend response / request shape
2. update frontend API wrapper in `static/src/api`
3. update shared TS types in `static/src/types/api/api.d.ts`
4. update UI usage
5. update `docs/api/route-index.md` if the public route surface or permission boundary changed

Rules:

- prefer additive changes over breaking changes
- preserve field names unless there is a clear bug
- use explicit JSON field names
- do not silently rename backend fields on the frontend

## 4. Localization Standard

All user-facing text must be localized.

Required:

- no hard-coded Chinese or English strings in views, dialogs, tables, empty states, buttons, or toast messages
- add entries to both:
  - `static/src/locales/langs/zh.json`
  - `static/src/locales/langs/en.json`
- prefer existing namespaces before creating new ones

Allowed exceptions:

- developer comments
- internal debug logs
- clearly isolated demo / seed content that is not user-facing

Preferred pattern:

- template: `$t('namespace.key')`
- script: `const { t } = useI18n()` then `t('namespace.key')`

## 5. Backend Standards

### 5.1 Responses

Use the existing unified response helpers in `server/pkg/response`. Do not invent per-handler response envelopes.

### 5.2 Authorization

- coarse access control belongs in router / middleware
- fine-grained ownership and business-role checks belong in service
- do not rely on frontend-only authorization

### 5.3 Persistence

- repositories should query only what they need
- if the frontend needs enriched rows, prefer dedicated DTO / response models instead of polluting base persistence models
- keep joins explicit and readable

### 5.4 External Integrations

- ESI / SSO calls belong in service or `pkg/eve`, not in handlers or repositories
- isolate retry / timeout behavior
- log failures with actionable context

## 6. Frontend Standards

### 6.1 Page Composition

- keep pages thin
- extract repeated UI into components
- extract repeated data behavior into hooks when it is reused
- prefer computed / render helpers over duplicated inline formatting logic

### 6.2 State

- page-local state stays in the page
- shared cross-page state goes to Pinia
- do not put server-derived state into global store unless multiple routes need it

### 6.3 Tables and Forms

- use existing shared patterns (`ArtTable`, `ArtTableHeader`, `useTable`, shared dialogs) when possible
- keep column labels localized
- keep search placeholders localized
- keep validation messages localized

### 6.4 Auth Pages

- the supported product auth flow is EVE SSO login + callback
- do not document or extend username / password auth as a supported flow unless explicitly requested
- if you touch legacy auth template pages, call out clearly whether the change is for active behavior or just cleanup

## 7. Type Safety Standard

- do not use `any` unless there is no practical alternative
- prefer existing `Api.*` types
- if a response is a special case, create a named interface or dedicated type
- keep backend and frontend field naming aligned exactly

## 8. Change Management Rules

Before editing:

- inspect the surrounding module first
- follow existing patterns in that slice of the codebase
- do not refactor unrelated areas opportunistically

When editing:

- keep changes scoped
- preserve backward compatibility where feasible
- avoid hidden coupling
- prefer explicit names over short clever ones

After editing:

- validate the exact layer you changed
- if you changed contracts, validate both backend and frontend
- update docs when current behavior or route surfaces changed

## 9. Verification Checklist

There are no repo-wide helper scripts such as `./scripts/setup-local.sh` or `./scripts/run-local-checks.sh` in the current tree. Use direct layer checks instead.

Recommended validation:

```bash
cd server && go test ./...
cd server && go build ./...
cd static && pnpm lint .
cd static && pnpm build
cd static && pnpm exec vue-tsc --noEmit
```

Minimum expectation:

- backend changes: `go test` and `go build`
- frontend changes: `pnpm lint .` and `vue-tsc --noEmit`
- cross-contract changes: validate both

## 10. Documentation Rules

Update documentation when behavior changes materially.

Usually relevant files:

- `README.md` for setup or product-facing workflow changes
- `docs/README.md` for documentation structure changes
- `docs/architecture/*` for current architecture or runtime workflow changes
- `docs/api/route-index.md` for API route / permission surface changes
- `docs/features/current/*` for current module behavior changes
- `AGENTS.md` for engineering standards

Notes:

- The maintained documentation tree lives under `docs/`

## 11. Anti-Patterns

Avoid these:

- hard-coded UI strings
- handlers with business logic
- repositories with authorization logic
- views with direct HTTP calls
- duplicated API types
- silently renamed fields across backend / frontend
- unrelated refactors mixed with feature fixes
- adding new patterns when an established repo pattern already exists
- N+1 database queries
- leaking internal errors to clients
- direct ESI calls from handlers
- business logic inside Vue views
- adding global store state unnecessarily

## 12. Preferred Change Pattern

For most feature work in this repository:

1. inspect the existing backend and frontend slice
2. identify the contract boundary
3. make the backend change
4. sync frontend API / types
5. update the UI
6. add localization entries
7. run targeted verification
8. update docs if routes, contracts, or supported behavior changed
