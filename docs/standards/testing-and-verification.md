---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-03-24
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/guides/regression-test-plan.md
  - server/go.mod
  - static/package.json
---

# Testing and Verification Standard

## Scope

Applies to all backend, frontend, contract, repository, hook, handler, and service changes in this repository.

## Core Rules

- Verification has two layers: build-level verification and behavior-level verification.
- `build`, `lint`, and `typecheck` do not replace regression testing.
- Any change that fixes a bug, changes a contract, or modifies non-trivial logic must be evaluated for regression coverage.
- Tests must exercise the real logic that changed. Do not duplicate a second implementation inside the test.

## Default Commands

### Backend

- `cd server && go test ./...`
- `cd server && go build ./...`

### Frontend

- `cd static && pnpm lint .`
- `cd static && pnpm exec vue-tsc --noEmit`
- `cd static && pnpm test:unit`

## Required Rules

- Bug fixes must add or update a regression test when the behavior can reasonably be tested.
- Backend logic changes should add `_test.go` coverage in the corresponding Go package.
- Repository changes involving query composition, mapping merges, filtering rules, branch selection, or fallback selection must add Go tests that cover the critical branches.
- Pure frontend helper or hook logic should add `pnpm test:unit` coverage.
- API contract changes must add behavior-level coverage on at least one affected side and must validate both backend and frontend.
- Any test command added to documentation must be runnable as written.

## Test Selection Guidance

### Prefer backend Go tests for:

- pure functions
- permission checks
- normalization logic
- repository branch logic
- fallback selection
- query composition helpers
- SQL fragment generation helpers

### Prefer frontend unit tests for:

- pure helpers
- pure hooks
- deterministic state transitions
- deduplication logic
- merge logic
- fallback logic
- namespace or mapping helpers

### Do not introduce heavy test infrastructure for:

- one-off minor regressions
- behavior that can be covered by extracting and testing a pure helper instead

If a frontend behavior truly requires full component mounting, browser APIs, or heavy mocking, first evaluate whether that test style should be introduced as a reusable repository pattern.

## Allowed Exceptions

New tests may be omitted only when the reason is stated explicitly in the change summary or review notes and one of the following applies:

- documentation-only changes
- formatting-only changes
- clearly behavior-preserving renames
- missing infrastructure makes temporary setup cost disproportionate to the change
- external dependencies or runtime conditions make reliable repository-local testing impractical

## Minimum Verification by Change Type

- backend-only change -> run backend test and build commands
- frontend-only change -> run frontend lint and typecheck, plus unit tests when relevant
- contract change -> validate both backend and frontend
- bug fix -> add or update regression coverage unless explicitly justified
- documentation-only change -> no code-level verification is required unless commands or executable examples changed

## Repository Notes

- Frontend unit testing in this repository is intentionally lightweight and best suited to pure logic.
- `static/src/types/import/auto-imports.d.ts` and `static/src/types/import/components.d.ts` are retained repository artifacts so clean checkouts can pass lint and typecheck.
- `static/.auto-import.json` is a local development helper and must not be required for CI linting.
- See `docs/guides/testing-guide.md` for naming, placement, and implementation guidance.
- See `docs/guides/regression-test-plan.md` for the repository-wide incremental regression strategy.

## Pre-Completion Checklist

Before considering a change complete, verify:

- Did the change fix a bug, change a contract, alter fallback behavior, or modify non-trivial branching?
- If yes, was regression coverage added or updated?
- Were the minimum required verification commands run?
- If tests were skipped, is the reason stated clearly?
- If a new documented test command was added, was it actually run locally?