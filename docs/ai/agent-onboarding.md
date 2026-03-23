---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-24
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/README.md
  - docs/ai/harness-principles.md
---

# AI Agent Onboarding

## Purpose

This guide helps AI agents reach the correct repository context quickly and make conservative, maintainable decisions when code and documentation do not fully align.

It is an onboarding and routing guide, not the primary rule source.

## Start Here

Before doing any work:

1. Read your agent entry point (`AGENTS.md` or `CLAUDE.md`) — both delegate to `docs/ai/repo-rules.md`.
2. Read `docs/README.md`.
3. Identify the change type.
4. Read the relevant architecture, API, feature, and standard documents before editing code.

For the harness model behind these rules, see `docs/ai/harness-principles.md`.

## Trust Hierarchy

See "Authority Order" in `docs/ai/repo-rules.md` for the canonical trust hierarchy.

Additional routing rules for agents:

- `docs/templates/` are not current-state authority
- subdirectory `README.md` files are local implementation notes, not repository-wide rule sources
- legacy compatibility files are not authoritative unless the current task explicitly targets them

## Context Boundaries

See "Context Boundaries" in `docs/ai/repo-rules.md` for the canonical definition.

In short: reason only from committed repository artifacts and information explicitly provided by the user in the current session.

## Minimum Reading Order by Change Type

### Backend or API Change

Read in this order:

1. `docs/ai/repo-rules.md` (loaded automatically via your agent entry point)
2. `docs/README.md`
3. `docs/architecture/overview.md`
4. `docs/architecture/module-map.md`
5. `docs/architecture/auth-and-permissions.md`
6. `docs/api/conventions.md`
7. `docs/api/route-index.md`
8. the relevant feature document

### Frontend Page, Route, or Permission Change

Read in this order:

1. `docs/ai/repo-rules.md` (loaded automatically via your agent entry point)
2. `docs/README.md`
3. `docs/architecture/module-map.md`
4. `docs/architecture/routing-and-menus.md`
5. `docs/standards/frontend-table-pages.md` when the page is a standard table page
6. the relevant feature document

### ESI, SSO, or CCP Data Sync Change

Read in this order:

1. `docs/ai/repo-rules.md` (loaded automatically via your agent entry point)
2. `docs/README.md`
3. `docs/architecture/overview.md`
4. `docs/architecture/module-map.md`
5. `docs/architecture/runtime-and-startup.md`
6. `docs/features/current/auth-and-characters.md`
7. `docs/features/current/esi-refresh.md`
8. `docs/guides/adding-esi-feature.md`

Read a local directory `README.md` under `server/pkg/eve/esi/` only after the task is clearly in that area.

## Conflict Resolution

See "Authority Order" and "Code-vs-Docs Rule" in `docs/ai/repo-rules.md` for canonical conflict resolution.

Key routing rules for agents:

- Do not use templates or local directory `README.md` files to override canonical repository rules.
- Do not revert working user behavior only to satisfy stale documentation.

## Before Editing Code

Before making changes:

- read the surrounding module code, not just a single file
- identify the relevant feature, API, and architecture documents
- identify whether the task affects standards, current-state docs, API docs, feature docs, or draft proposals
- if the content is only a future idea, do not rewrite current-state documents to describe it as implemented

## Minimum Documentation Updates After Changes

Update the following when applicable:

- behavior changed -> update the relevant feature document
- route surface or permission boundary changed -> update `docs/api/route-index.md`
- runtime or startup behavior changed -> update `docs/architecture/runtime-and-startup.md`
- repository-wide rule or engineering standard changed -> update `docs/ai/repo-rules.md` or the relevant file under `docs/standards/`

## Required Behavior

- Read `docs/ai/repo-rules.md` (via your agent entry point) before starting work.
- Read the relevant code and feature documentation before changing code.
- Follow the completion and verification protocol in `docs/standards/pre-completion-checklist.md`.
- When blocked, stop and reassess instead of retrying the same approach repeatedly.
- When code and docs disagree, determine whether the drift is in the code or in the documentation.

## Prohibited Behavior

See "Anti-Drift Rules" and "Guardrails" in `docs/ai/repo-rules.md` for the full list.

Key prohibitions for agent onboarding:

- do not create a second shadow documentation tree
- do not write planned or future behavior into current-state documents
- do not treat template files or local directory `README.md` files as repository-wide source of truth
- do not keep editing the same file repeatedly without progress
- do not introduce patterns that conflict with established repository conventions

## Debugging Guidance

When debugging, use the repository debugging workflow in `docs/guides/debugging-guide.md`.

Default approach:

1. classify the problem
2. locate the affected layer
3. reproduce it with the smallest useful scope
4. fix the root cause
5. add regression protection where appropriate

## Verification Reference

At the end of the task, use the relevant checklist in `docs/standards/pre-completion-checklist.md`.

For verification commands, see `docs/standards/testing-and-verification.md`.