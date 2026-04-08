---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-04-09
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/architecture/module-map.md
---

# Debugging Guide

## Purpose

This guide helps developers and agents systematically diagnose issues rather than guessing. It applies to both runtime bugs and build/test failures.

## General Debugging Protocol

### Step 1: Classify the Problem

Before investigating, classify the issue:

| Category | Symptoms | Start Looking At |
| --- | --- | --- |
| Build failure | `go build` or `pnpm build` fails | error message, recent changes |
| Type error | `vue-tsc --noEmit` fails | `api.d.ts`, API wrappers, component props |
| Lint error | `pnpm lint .` fails | ESLint output, auto-fixable vs manual |
| Test failure | `go test` or `pnpm test:unit` fails | test output, recent changes to tested code |
| Runtime backend | HTTP 500, wrong response, panic | server logs, handler → service → repository |
| Runtime frontend | wrong data displayed, missing UI | browser console, API response vs type |
| Permission error | 403, page/button missing | route protection, route meta, role assignments |
| Data inconsistency | wrong values in UI | repository query, join conditions, fallbacks |

### Step 2: Locate the Layer

Use the module map (`docs/architecture/module-map.md`) to identify which files are involved.

**Backend flow:** `router → middleware → handler → service → repository → model`

Start from the symptom and trace toward the root:

- Wrong HTTP response? Start at handler, check what service returns
- Wrong data? Start at repository, check query/joins
- Wrong authorization? Start at middleware/router, check role requirements
- Wrong business logic? Start at service

**Frontend flow:** `view → api → backend`

- Wrong display? Start at view template, check data binding
- Wrong data? Start at API wrapper, check response type
- Missing/wrong text? Check i18n keys in zh.json / en.json
- Permission issue? Check route meta, v-auth, store permissions

### Step 3: Reproduce with Minimum Scope

- For backend: write a focused test that reproduces the issue
- For frontend: isolate the component/hook that produces the wrong result
- For cross-layer: determine which side is wrong first by checking the raw API response

### Step 4: Fix at the Root Cause

Do not patch symptoms. Common mistakes:

| Symptom | Wrong Fix | Right Fix |
| --- | --- | --- |
| Wrong name displayed | Change frontend fallback text | Fix repository query/join |
| 403 on valid user | Remove permission check | Fix role assignment or middleware |
| Type error after API change | Cast to `any` | Update `api.d.ts` types |
| Duplicate data | Add `distinct` to query | Fix the join that causes duplication |

### Step 5: Add Regression Test

After fixing, add a test that would have caught this bug. See `docs/standards/regression-test-plan.md` for the decision matrix on which layer to test.

## Common Issue Patterns

### Join Query Issues

**Symptoms:** wrong data, duplicate rows, NULL fields

**Investigation:**

1. Read the repository query
2. Check all JOIN conditions for correctness
3. Check for ambiguous column names (e.g., `id`, `status`, `deleted_at` exist in multiple tables)
4. Check for missing `WHERE table.deleted_at IS NULL` on soft-deleted tables

**Prevention:** use explicit table-qualified column names in SELECT and WHERE clauses.

### API Contract Mismatch

**Symptoms:** frontend type errors, wrong data in UI, `undefined` fields

**Investigation:**

1. Compare backend handler/service response struct with `api.d.ts` type
2. Check JSON tags on backend structs match frontend field names
3. Check if a recent backend change wasn't propagated to frontend

**Prevention:** update all 5 layers when changing an endpoint (see "API Change Order" in `docs/ai/repo-rules.md`).

### Permission / Menu Visibility Issues

**Symptoms:** pages not visible, buttons missing, 403 errors

**Investigation:**

1. Check backend route protection in `router.go`
2. Check frontend route meta (`meta.login`, `meta.roles`)
3. Check `v-auth` directives on buttons
4. Check user's actual roles via `/api/v1/me`

**Prevention:** modify backend route protection, frontend route metadata, and button permission touchpoints together (see "Routing, Menu, and Permission Changes" in `docs/ai/repo-rules.md`).

### Localization Key Missing

**Symptoms:** raw key strings shown in UI (e.g., `operation.fleetList`)

**Investigation:**

1. Search for the key in both `zh.json` and `en.json`
2. Check for typos in the key path
3. Check that the namespace exists

**Prevention:** always add i18n entries to both language files in the same change.

### ESI / SSO Integration Issues

**Symptoms:** token refresh failures, missing data, scope errors

**Investigation:**

1. Check `server/config/config.go` ESI URL configuration
2. Check ESI refresh task in `server/pkg/eve/esi/`
3. Check token validity and scope requirements
4. Check ESI rate limiting (HTTP 420 responses in logs)

**Prevention:** never hardcode ESI URLs. Use `global.Config.EveSSO.ESIBaseURL`.

## Build Failure Quick Reference

### `go build` fails

1. Read the error message — it usually points to the exact file and line
2. Check for missing imports, type mismatches, or undefined variables
3. Run `go mod tidy` if import errors
4. Check if model changes broke repository/service signatures

### `vue-tsc --noEmit` fails

1. Read the type error — file, line, expected vs actual type
2. Common cause: API response type changed but `api.d.ts` wasn't updated
3. Common cause: component prop types changed
4. Check auto-generated import files are present

### `pnpm lint .` fails

1. Read ESLint output
2. Many issues are auto-fixable: `pnpm lint . --fix`
3. For non-auto-fixable: read the rule name and fix manually
4. Do not disable rules without justification

### `go test ./...` fails

1. Read the test output to identify which test failed
2. Check if the test expectation is still correct after your change
3. If behavior intentionally changed, update the test
4. If behavior shouldn't have changed, your code has a regression

## When to Ask for Help

If after following this protocol you are still stuck:

- You've traced the issue through multiple layers without finding the root cause
- The fix requires changes to infrastructure you don't understand
- The issue involves external systems (ESI, CCP) that may have changed
- You're unsure whether the current behavior is intentional or a bug

Surface the blocker clearly: what you tried, what you found, and where you got stuck.
