---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-26
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/guides/regression-test-plan.md
  - server/go.mod
  - static/package.json
---

# Testing Guide

## Purpose

This guide is for fast test-placement and test-shape decisions.

Repository policy lives in `docs/standards/testing-and-verification.md`.

## Default Locations

### Backend

- Put Go tests in the same package as the code under test using `*_test.go`.
- Prefer tests near:
  - `server/internal/service/`
  - `server/internal/handler/`
  - `server/internal/repository/`

### Frontend

- Put pure helper and hook tests next to the file under test when practical.
- Use `cd static && pnpm test:unit` for current frontend unit coverage.

## Quick Heuristics

- Test the layer that owns the real behavior.
- For backend business rules, prefer Go tests in service, handler, repository, or helper packages.
- For frontend deterministic logic, prefer tests around helpers, hooks, state transitions, request mapping, and response merge logic.
- If one side is mostly wiring and the other side owns the real branching, add behavior coverage at the root-cause layer and use build-level verification on the wiring side.
- Read nearby tests before changing an existing feature. They often define the current contract faster than code alone.

## Prefer Testing

### Backend

- service permission checks
- normalization logic
- filter and enum handling
- repository branch and fallback behavior
- handler-boundary contract logic

### Frontend

- pure helpers
- pure hooks
- deterministic state transitions
- deduplication and merge logic
- request-parameter transformation

## Avoid By Default

- heavy component-test infrastructure for a small logic change
- temporary complex database setups for low-risk branches
- tests that only mirror implementation details without checking stable behavior

## Naming

### Go

- Prefer `TestFunctionNameScenario`
- Examples:
  - `TestParseEFTHeader`
  - `TestNormalizeSkillPlanName`

### Frontend

- Prefer behavior-based names
- Examples:
  - `mergeNamesResponse keeps namespace-specific values`
  - `buildPendingRequest keeps type and solar_system ids separate`

## When Stuck

- Extract a small pure unit instead of skipping tests entirely.
- Add the smallest test that locks down the changed behavior.
- If new tests are not practical, document the reason in the change summary.
