---
status: active
doc_type: agent-rules
owner: engineering
last_reviewed: 2026-03-24
source_of_truth:
  - AGENTS.md
  - CLAUDE.md
  - docs/README.md
---

# Repository Rules

## Scope

This document defines repository-wide agent rules, authority order, and completion requirements for AmiyaEden.

Detailed standards, feature behavior, and implementation guidance live under `docs/`.

## Authority and Context

### Authority Order

When sources conflict, authority descends in this order:

1. this file
2. `docs/standards/*.md`
3. `docs/architecture/*.md`
4. `docs/api/*.md`
5. `docs/features/current/*.md`
6. `docs/guides/*.md`
7. `docs/specs/draft/*.md`

### Context Boundaries

Authoritative context is limited to committed repository artifacts:

- code
- docs
- schemas
- config
- migrations
- tests
- checked-in generated types

Treat Slack, Google Docs, verbal decisions, and other uncommitted knowledge as non-authoritative unless the user provides them in the current session.

### Code-vs-Docs Rule

Code is the current implementation.

If code and docs conflict:

1. inspect the relevant code path
2. determine whether the doc is stale or the code is wrong
3. update the stale doc if behavior is intentional
4. only change code to match docs when the user explicitly wants that outcome or the docs clearly reflect the intended requirement

## Project Intent

`AmiyaEden` is a full-stack EVE Online operations platform with:

- Go backend in `server/`
- Vue 3 + TypeScript frontend in `static/`
- RBAC roles, menus, and button permissions
- dynamic menu and routing support
- ESI / SSO integrations
- typed frontend API contracts

The supported authentication flow is EVE SSO. Legacy auth-related pages may still exist, but they are not current product requirements unless the user explicitly asks to work on them.

## Non-Negotiable Rules

1. **Layering is law**
   - Backend: `router -> middleware -> handler -> service -> repository -> model`
   - Frontend: `view -> api -> backend`

2. **Contracts are synchronized**
   - API changes must update backend, frontend API wrappers, TypeScript types, UI usage, and relevant docs in the same change

3. **All user-facing text is localized**
   - No hardcoded UI strings
   - Update both `zh.json` and `en.json`

4. **Type safety over convenience**
   - Do not introduce `any` unless clearly justified
   - Prefer existing shared types

5. **Business logic belongs in services**
   - Not in handlers
   - Not in repositories
   - Not in Vue views

6. **Permissions are enforced server-side**
   - Frontend permission logic is UX only

7. **Changes are scoped**
   - Do not mix unrelated refactors into feature work or bug fixes

8. **Bug fixes require regression protection**
   - Add a regression test unless impractical, and justify any omission

9. **Docs change with behavior**
   - Update relevant docs in the same change when behavior, contracts, routes, or workflows change

10. **Prefer established patterns**
   - Follow repository conventions before introducing new abstractions

## Documentation Routing

Start here when working in unfamiliar areas:

- `docs/README.md`
- `docs/architecture/module-map.md`
- `docs/standards/dependency-layering.md`
- `docs/api/conventions.md`
- `docs/standards/testing-and-verification.md`
- `docs/standards/pre-completion-checklist.md`

### Feature Specs

Feature behavior specs live under `docs/features/current/`.

Before changing feature behavior:

1. read the relevant feature spec
2. read `docs/architecture/module-map.md`
3. inspect the current implementation in code

If no feature spec exists, use the current code and standards as the source of truth, then add or update the relevant feature doc when behavior changes materially.

## Architecture and Contract Rules

### Layer Responsibilities

- **handler**: transport only
- **service**: business rules and orchestration
- **repository**: data access only
- **model**: persistence and contract structures

### Routing, Menu, and Permission Changes

When changing roles, menus, routes, or button permissions, keep the following aligned:

- backend route protection
- menu seeds in `server/internal/model/menu.go`
- frontend route metadata
- button permission usage such as `v-auth`

### API Change Order

When changing an endpoint, update in this order:

1. backend request or response shape
2. frontend API wrapper in `static/src/api`
3. shared TypeScript types in `static/src/types/api/api.d.ts`
4. UI usage
5. `docs/api/route-index.md` if the route surface or permission boundary changed

Do not allow contract drift across backend, API wrappers, shared types, and UI usage.

## Repository-Specific Rules

### Backend

- Use `server/pkg/response` helpers
- Keep coarse auth in router or middleware
- Keep fine-grained auth in services
- Keep ESI and SSO logic in services or `pkg/eve`
- Avoid leaking internal errors
- Avoid N+1 queries

### Frontend

- Keep page components thin
- Extract repeated UI into components
- Extract repeated logic into hooks or composables
- Keep local state local; use Pinia only when state is genuinely cross-page
- Use established table and form patterns
- Do not make direct HTTP calls in views
- Do not extend legacy username/password auth flows unless explicitly requested

### Localization

Use the existing i18n patterns and update both `zh.json` and `en.json` in the same change.

## Verification and Completion

Use:

- `docs/standards/testing-and-verification.md`
- `docs/standards/pre-completion-checklist.md`

Minimum expectations:

- backend changes -> `cd server && go test ./...` and `go build ./...`
- frontend changes -> `cd static && pnpm lint .` and `pnpm exec vue-tsc --noEmit`
- contract changes -> validate both backend and frontend
- bug fixes -> add a regression test or state why not
- behavior changes -> update relevant docs

Before completion:

1. re-check scope against the request
2. confirm no unrelated refactors were introduced
3. run relevant verification
4. confirm contract synchronization where applicable
5. confirm localization coverage for user-facing text
6. update affected docs
7. note any remaining risk or unverified area

## Guardrails

### When Blocked or Looping

If progress stalls, re-read the relevant docs, inspect the current implementation again, change approach, and surface the blocker if still unresolved.

### Anti-Drift Rules

- do not introduce patterns that contradict repository conventions
- do not broaden scope without clear reason
- do not add one-off abstractions with no demonstrated reuse
- do not leave commented-out code, placeholder files, or context-free TODOs

### Divergence Rule

When several files follow an established pattern and one diverges, assume the diverging file needs justification before copying it.

## Quick Start

See `docs/guides/local-development.md` for full setup instructions.

For verification commands, see `docs/standards/testing-and-verification.md`.

## Maintenance

Update this document only when repository-wide agent rules change.

When updating it:

- keep it concise and policy-focused
- move procedural detail into `docs/standards/`
- keep feature behavior in `docs/features/current/`
- update `last_reviewed`