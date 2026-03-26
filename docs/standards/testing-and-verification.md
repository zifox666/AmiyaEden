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

Applies to backend, frontend, contract, repository, hook, handler, and service changes.

## Required Tool Versions

| Tool | Version | How to install |
|------|---------|----------------|
| golangci-lint | v2.11.4 | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.11.4` |
| pnpm | 10.32.1 | `npm install -g pnpm@10.32.1` |
| Node.js | 24 | see `.nvmrc` at repo root |

- CI pins `golangci-lint` in `.github/workflows/verify-ci.yaml`.
- Frontend packages are locked by `static/pnpm-lock.yaml`; use `pnpm install --frozen-lockfile`.

## Default Commands

Single canonical source for verification commands.

### Backend

- `cd server && golangci-lint run ./...`
- `cd server && go test ./...`
- `cd server && go build ./...`

### Frontend

- `cd static && pnpm lint .`
- `cd static && pnpm exec vue-tsc --noEmit`
- `cd static && pnpm test:unit`
- `cd static && pnpm build`

## Rules

- `build`, `lint`, and `typecheck` do not replace behavior-level coverage.
- Tests must exercise the real changed logic; do not reimplement production logic in the test.
- New features must add or update relevant automated coverage when reasonably testable.
- Existing feature changes must review and update nearby tests when covered behavior or contracts change.
- Bug fixes must add or update regression coverage when reasonably testable.
- Backend logic changes should add `_test.go` coverage in the same Go package.
- Repository branch, filter, merge, query, and fallback logic must add Go tests for critical branches.
- Pure frontend helper or hook logic should add `pnpm test:unit` coverage.
- API contract changes must validate both backend and frontend and add behavior-level coverage on at least one affected side.
- Any documented test command must be runnable as written.

## Test Choice

- Prefer backend Go tests for service rules, normalization, permission checks, query helpers, repository branching, and fallback logic.
- Prefer frontend unit tests for pure helpers, pure hooks, deterministic state transitions, merge logic, fallback logic, and request mapping.
- Avoid heavy test infrastructure for small logic changes when a lightweight unit can cover the behavior instead.

## Allowed Exceptions

Tests may be omitted only if the reason is stated explicitly and one of these applies:
- documentation-only changes
- formatting-only changes
- clearly behavior-preserving renames
- missing infrastructure makes temporary setup cost disproportionate to the change
- external dependencies or runtime conditions make reliable repository-local testing impractical

## Minimum Verification

- backend-only change -> run backend test and build commands
- frontend-only change -> run frontend lint and typecheck, plus unit tests when relevant
- contract change -> validate both backend and frontend
- new feature -> add or update relevant automated coverage for the new behavior unless an allowed exception is stated explicitly
- bug fix -> add or update regression coverage unless explicitly justified
- existing feature behavior change -> review and update existing tests where needed, and add coverage for new or changed behavior unless an allowed exception is stated explicitly
- documentation-only change -> no code-level verification is required unless commands or executable examples changed

## Repository Notes

- Frontend unit testing is intentionally lightweight and best suited to pure logic.
- `static/src/types/import/auto-imports.d.ts` and `static/src/types/import/components.d.ts` are retained so clean checkouts pass lint and typecheck.
- `static/.auto-import.json` must not be required for CI linting.
- See `docs/guides/testing-guide.md` for placement and implementation guidance.
- See `docs/guides/regression-test-plan.md` for incremental regression planning.

## Completion Check

- New feature or changed feature behavior: was relevant coverage added or updated?
- Bug fix, contract change, fallback change, or non-trivial branching change: was regression coverage added or updated?
- Were the minimum required commands run?
- If tests were skipped, is the reason stated clearly?
- If a new documented test command was added, was it run locally?
