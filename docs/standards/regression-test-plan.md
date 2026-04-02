---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-04-02
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/standards/testing-and-verification.md
  - docs/guides/testing-guide.md
  - server/
  - static/
---

# Regression Test Plan

## Purpose

This document is not a new mandatory rule source. It translates the existing testing standard into an incremental implementation plan, answering three practical questions:

- Which regression tests should this repository prioritize?
- What is the minimum test needed to prevent each class of bug from recurring?
- How can regression coverage improve incrementally without a one-time test infrastructure overhaul?

Intended audience:

- Developers fixing bugs
- Developers refactoring modules
- PR reviewers
- Documentation and engineering standards maintainers

## Goals

- Prevent "fixed once, broke again" issues by catching them locally with tests
- Prioritize test coverage for high-risk boundaries: permissions, fallbacks, query joins, and contracts
- Allow new tests to fit the current code structure without requiring large new infrastructure first
- Let each module gradually accumulate a stable set of regression examples

## Non-Goals

- Do not require backfilling all module tests at once
- Do not require a full e2e framework for a single small bug fix
- Do not treat build / lint / typecheck as behavior-level regression tests
- Do not push high-maintenance UI component snapshot tests at this stage

## Core Strategy

Default to "test at the layer closest to the bug":

1. If the bug is in pure logic, normalization, or permission checks, prefer a service / helper unit test.
2. If the bug is in a repository query, join, fallback, or field mapping, prefer a repository regression test.
3. If the bug is in an API contract, response shaping, or pagination envelope, prefer a handler / API contract test.
4. If the bug is in a frontend pure helper, filter parameter transform, or name fallback, prefer a frontend unit test.
5. If the bug surfaces only in the page assembly layer but the root cause is a backend contract issue, add the backend test first, then verify the frontend build.

Do not default to the heaviest test layer. Choose the smallest, most stable layer that locks down the real risk.

## Risk Layers and Recommended Test Types

| Risk Type | Common Examples | Minimum Recommended Test |
| --- | --- | --- |
| Permission boundary | admin editing admin, guest accessing login page | service unit test |
| Input validation / normalization | nickname, QQ, Discord, time range, enum correction | service or helper unit test |
| Repository query join | column ambiguity after join, missing filter, wrong sort | repository SQL / query-shape test |
| Repository fallback / merge | nickname falling back to character name, role list falling back to guest | repository behavior test |
| API contract | field name change, roles[] vs role difference, pagination structure | handler or API contract test |
| Frontend pure logic | filter parameter transform, fallback text, table helper | `pnpm test:unit` |
| Localization regression | missing key, page displaying raw key | JSON validation + manual verification on page changes |
| Page assembly error | wrong column mapping, wrong field binding, wrong button condition | prefer helper / contract test; add lightweight frontend test if needed |

## Current Repository Priorities

First priority modules:

- `operation`: fleets, fleet-detail, pap, fleet-configs

  Fleet-configs bug fixes should prefer regression coverage for EFT parse / rebuild round trips and equipment-setting preservation. In particular, verify that settings are preserved only when `flag + type_id + quantity` remain unchanged, and that changed or removed items reset to defaults.
- `system`: user, role, auto-role, pap, webhook
- `auth-and-characters`: `/api/v1/me`, character binding, profile completion

Reasons:

- These modules involve permissions, query joins, frontend-backend contracts, and fallback display simultaneously
- Join query regressions and display field fallback regressions have already occurred recently
- These modules have high daily usage impact, and their bugs are typically not caught at compile time

Second priority modules:

- `srp`
- `commerce`
- `info-and-reporting`
- `skill-planning`

Third priority modules:

- Documentation, static configuration, low-risk read-only pages

## Phased Implementation

### Phase 1: Lock Down New Bugs

Goal: from now on, all new bug fixes should include a minimum regression test.

Requirements:

- For every bug fix, first ask "which layer is closest to the root cause?"
- When reasonably testable, a regression test targeting that bug is required
- If infrastructure is currently missing, at minimum add a query-shape / helper / service level test

Completion criteria:

- New bug fixes no longer rely solely on `go build` or `vue-tsc`
- Recent regression points begin to have corresponding tests

Suggested early examples:

- Fleet list FC nickname fallback
- Column ambiguity after joins (`deleted_at`, `status`, `id`)
- User list role fallback and sorting
- admin / super_admin protection logic (super_admin is managed only via config file; API cannot assign / modify / delete it)
- `/api/v1/me` profile completion and contact info uniqueness

### Phase 2: Add Module-Level High-Frequency Regression Points

Goal: build a stable "protection belt" for frequently modified modules.

Each high-priority module should have at least:

- 2 to 5 service / helper regression tests
- 1 to 3 repository regression tests
- 1 key contract test

Module suggestions:

### Operation

- `fleet list` query join and FC display fallback
- PAP log display field fallback logic
- auto SRP mode normalization
- fleet permission checks: `fc` / `admin` / `super_admin`

### Administration

- User profile update validation and uniqueness
- Protected admin accounts cannot be modified / deleted by regular admins
- super_admin role cannot be granted, modified, or deleted via API
- super_admin users cannot be deleted via API
- super_admin role syncs automatically from config file on login
- Role list `roles[]` and legacy `role` fallback
- `GET /system/basic-config` returns only fixed system identifiers with no corresponding write endpoint
- auto-role `Director -> admin` rule only accepts corp role signals from Fuxi Legion (`98185110`)
- `allow_corporations` always retains `98185110` on save and read

### Auth And Characters

- Profile completeness check
- Character binding / main character switch permission and input validation
- `guest` to `user` boundary behavior

### Phase 3: Build Shared Test Fixtures

Goal: reduce the repeated environment setup cost for each new test.

Suggested additions (not required all at once):

- Backend repository dry-run GORM helper
- Backend handler test helper
- Frontend locale JSON validation helper
- Frontend API contract mock helper

Notes:

- The current repository is already suited for dry-run SQL / schema mapping tests
- If repository integration tests grow significantly, consider a unified test database fixture later
- Do not build a complex test platform preemptively for hypothetical future use

## Specific Test Patterns

### 1. Repository Query-Shape Test

Applicable when:

- join changes
- SQL select / where / order / fallback changes
- new computed fields

Purpose:

- Ensure critical SQL fragments are present
- Ensure column ambiguity does not recur
- Ensure fallback expressions are preserved

Examples:

- `fleet.deleted_at IS NULL`
- `LEFT JOIN "user"`
- `COALESCE(NULLIF("user".nickname, ''), fleet.fc_character_name)`

These tests are particularly suited to this repository because:

- They run fast
- They do not depend on a real database
- They catch the join regressions that have occurred recently

### 2. DTO / Schema Mapping Test

Applicable when:

- query adds a new alias field
- a special DTO field is used only in the response and not persisted
- a GORM tag typo causes the query to return data that fails to map

Purpose:

- Ensure query aliases actually scan into the DTO
- Ensure field names align with JSON / DBName tags

### 3. Service Behavior Test

Applicable when:

- permission checks
- fallback rules
- input normalization
- uniqueness validation

Purpose:

- Lock down business rules
- Prevent policies from being scattered across handlers or pages without protection

### 4. Handler / API Contract Test

Applicable when:

- pagination structure changes
- field name changes
- response envelope changes
- permission boundary changes on important endpoints

Purpose:

- Prevent "compiles on backend but frontend contract is already broken"

### 5. Frontend Unit Test

Applicable when:

- pure helpers
- pure computations in hooks
- filter parameter transforms
- fallback text selection

Purpose:

- Protect frontend behavior with the lightest possible approach

Not prioritized at this stage:

- Heavy component mount tests for standard list pages
- End-to-end browser tests for simple text changes

## Bug Fix Minimum Regression Requirements

When fixing a bug, use this table directly:

| Bug Root Cause | Minimum Required |
| --- | --- |
| Service rule wrong | one service test |
| Repository query wrong | one repository regression test |
| Response field wrong | one handler / contract test |
| Frontend helper wrong | one frontend unit test |
| Multiple layers involved | root-cause layer test + build verification on another layer |

If a test cannot be added immediately, the change description must state:

- Why the test was not added now
- What infrastructure is missing
- Where the test should be added later

## Review Checklist

When reviewing a bug fix, at minimum ask:

1. Is the bug's root cause in the handler, service, repository, or frontend helper?
2. Does the new test actually lock down that root cause, not just surface behavior?
3. If someone later changes the same logic, will this test fail immediately?
4. Beyond build / lint / typecheck, is there behavior-level protection?

## Suggested Module Regression Checklists

The following are not one-time task lists but priority queues for incremental coverage.

### auth-and-characters

- `ProfileComplete()` stays consistent with frontend profile completion check
- QQ / Discord uniqueness
- `/api/v1/me` returns role and permission context

### operation

- Fleet list query and display fallback
- Fleet management permission checks
- PAP issuance preconditions
- auto SRP mode normalization and trigger conditions

### administration

- User list DTO no longer leaks legacy `role`
- User role sorting and fallback
- Admin cannot operate on protected accounts
- auto-role built-in shortcut rules vs title mapping distinction

### commerce

- Purchase limit rules
- Order status transitions
- Wallet transaction type and reference type mapping

### srp

- SRP application status transitions
- Fleet / KM association fallback
- Auto-approval vs manual approval boundary

## Command Reference

See `docs/standards/testing-and-verification.md` for verification commands.

## Documentation Maintenance

When a module begins to accumulate stable regression tests, update the corresponding feature doc to at minimum state:

- What key invariants the module currently has
- What high-risk protection points were recently added
- Which test layer protects those invariants

Do not copy specific test file lists into multiple documents for redundant maintenance. Testing strategy is governed by:

- `docs/standards/testing-and-verification.md`
- `docs/guides/testing-guide.md`
- This document
