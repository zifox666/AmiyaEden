---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-24
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/standards/testing-and-verification.md
  - docs/guides/regression-test-plan.md
  - server/go.mod
  - static/package.json
---

# Testing Guide

## Purpose

This guide complements `docs/standards/testing-and-verification.md` by answering practical questions:

- where tests usually belong
- when to add a unit test versus when build-level verification is enough
- what kinds of tests fit this repository well today

For planning which regression tests to add next and how to phase that work, see `docs/guides/regression-test-plan.md`.

## Default Test Locations

### Backend

- Package-level tests live in the same directory as the code under test and use `*_test.go`.
- For pure logic, permission checks, normalization, and mapping or merge logic, prefer tests in the corresponding Go package.
- Current high-value test coverage is mainly concentrated in:
  - `server/internal/service/`
  - `server/internal/handler/`
  - `server/internal/repository/`

### Frontend

- Prefer placing pure helper and pure hook tests next to the file under test.
- The current lightweight frontend test entry point is:
  - `cd static && pnpm test:unit`
- The current frontend test setup is best suited to:
  - pure functions
  - namespace, deduplication, fallback, and merge logic
  - state transitions that do not require a full DOM environment

## What This Repository Should Prefer to Test

### Backend

Prefer tests for:

- service-layer permission checks
- input normalization
- time-range handling, filter parsing, and enum mapping
- repository branch selection, merge logic, and fallback behavior
- pure helpers in handlers or contract-merge logic that can be isolated cleanly

### Frontend

Prefer tests for:

- pure helpers used by hooks
- API response merge logic
- name resolution, cache-key generation, and deduplication logic
- pure logic inside table column definitions or filter-parameter transformation

## What This Repository Should Not Prioritize Yet

Do not prioritize the following unless there is clear repeated value:

- introducing heavy component-test infrastructure for a very small logic change
- building a temporary complex test database setup only to cover a low-risk repository branch
- writing tests that only verify internal implementation details without verifying stable external behavior

## Test Naming Guidance

### Go

Prefer the pattern:

- `TestFunctionNameScenario`

Examples:

- `TestParseEFTHeader`
- `TestNormalizeSkillPlanName`
- `TestMergeGetNamesNamespacesPreservesNamespacesAndFlatFirstWins`

### Frontend

Prefer behavior-based names.

Examples:

- `mergeNamesResponse keeps namespace-specific values`
- `buildPendingRequest keeps type and solar_system ids separate`

## Test Writing Guidance

- Prefer testing public behavior or stable helper behavior. Do not copy production logic into the test.
- Use the smallest inputs that cover the most important branches. Do not try to exhaust every combination in one change.
- Lock down the bug first, then lock down the contract if relevant.
- For changes spanning backend and frontend, add behavior-level coverage on at least one side and perform build-level verification on the other side at minimum.
- If the change reveals that the repository lacks a suitable test entry point, prefer adding a lightweight reusable test path instead of leaving the verification process undocumented or hidden in discussion history.

## Common Commands

### Backend

```bash
cd server && go test ./...
cd server && go build ./...
```

### Frontend

```bash
cd static && pnpm lint .
cd static && pnpm exec vue-tsc --noEmit
cd static && pnpm test:unit
```

## When Verification without new tests is acceptable only when justified

The following cases usually do not require new tests, but the reason should be stated in the change summary:

- documentation-only changes
- formatting-only changes
- clearly behavior-preserving renames
- the repository does not currently have reasonable test infrastructure and temporary setup cost is clearly higher than the value of the change

## Review Questions

Before considering the work complete, ask:

- Did this change fix a bug?
- Did it change fallback behavior, merge logic, filtering, or permission boundaries?
- Did it change the backend/frontend contract?
- If yes, was regression coverage added or updated?
