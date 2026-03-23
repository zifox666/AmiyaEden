---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-03-24
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/standards/testing-and-verification.md
  - docs/standards/dependency-layering.md
---

# Pre-Completion Checklist

## Scope

Use this checklist before marking any task complete.

Skip only items that do not apply to the change. Do not skip items for convenience.

## Core Rules

- Completion requires both correct implementation and appropriate verification.
- `build`, `lint`, and `typecheck` do not replace regression testing where regression coverage is required.
- If a required check or test is skipped, the omission must be stated explicitly.

## Checklist by Change Type

### Backend-Only Change

- [ ] `cd server && go build ./...`
- [ ] `cd server && go test ./...`
- [ ] No layer violations were introduced
- [ ] If this is a bug fix, a regression test was added or updated
- [ ] If an API contract changed, frontend API wrappers and types were updated
- [ ] If a route was added or changed, `docs/api/route-index.md` was updated
- [ ] If behavior changed, the relevant feature doc was updated

### Frontend-Only Change

- [ ] `cd static && pnpm lint .`
- [ ] `cd static && pnpm exec vue-tsc --noEmit`
- [ ] If a pure helper or hook changed, `cd static && pnpm test:unit`
- [ ] No direct HTTP calls were added to views
- [ ] All new user-facing strings were added to both `zh.json` and `en.json`
- [ ] If behavior changed, the relevant feature doc was updated

### Cross-Contract Change

- [ ] `cd server && go build ./...`
- [ ] `cd server && go test ./...`
- [ ] `cd static && pnpm lint .`
- [ ] `cd static && pnpm exec vue-tsc --noEmit`
- [ ] If relevant, `cd static && pnpm test:unit`
- [ ] Frontend API wrapper was updated
- [ ] Shared TypeScript types were updated
- [ ] Backend response fields and frontend type fields match
- [ ] `docs/api/route-index.md` was updated if the route surface or permission boundary changed
- [ ] The relevant feature doc was updated if behavior changed

### Permission or Role Change

- [ ] All applicable items from Cross-Contract Change were completed
- [ ] Backend route protection was updated where required
- [ ] `server/internal/model/menu.go` was updated if menu seeds changed
- [ ] Frontend route metadata was updated where required
- [ ] Button permission usage such as `v-auth` was aligned
- [ ] Changes were validated against both frontend and backend menu modes if applicable
- [ ] `docs/architecture/auth-and-permissions.md` was updated if the permission model or behavior changed

### Documentation-Only Change

- [ ] Front matter was updated where required
- [ ] No stale references or broken cross-links were introduced
- [ ] Index documents were updated if required
- [ ] Current code was checked when the document describes current implementation

### New Feature or Module

- [ ] All applicable items from Cross-Contract Change were completed
- [ ] A feature doc was created under `docs/features/current/` if the feature has durable behavior
- [ ] Any relevant feature index was updated
- [ ] Localization was completed in both `zh.json` and `en.json`
- [ ] Menu seeds were added if required
- [ ] Backend and frontend routes were registered if required
- [ ] The change follows the existing module structure pattern
- [ ] At least one regression test covers key behavior, unless explicitly justified otherwise

## Test Decision Matrix

| change | minimum test expectation |
| --- | --- |
| service business logic | Go test in the same package |
| repository query, join, filter, or fallback logic | Go behavior or branch test |
| handler response shape or contract logic | Go handler-boundary or contract test |
| frontend pure helper or pure hook | `cd static && pnpm test:unit` |
| bug fix in any layer | regression test at the root-cause layer when practical |
| localization-only change | build-level verification only |
| documentation-only change | no code-level test required |

## If a Test Is Skipped

When a normally expected test is skipped:

1. state which test was skipped
2. state why it was skipped
3. state where the test should be added later, if applicable

Never skip a normally expected test without documenting the reason.

## Quick Commands

Backend:

- `cd server && go build ./...`
- `cd server && go test ./...`

Frontend:

- `cd static && pnpm lint .`
- `cd static && pnpm exec vue-tsc --noEmit`
- `cd static && pnpm test:unit`

Full stack:

- `cd server && go build ./... && go test ./... && cd ../static && pnpm lint . && pnpm exec vue-tsc --noEmit && pnpm test:unit`