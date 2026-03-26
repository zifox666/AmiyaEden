---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-26
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/README.md
---

# AI Agent Onboarding

## Purpose

Fast routing for agents. Repository rules live in `docs/ai/repo-rules.md`.

## Startup

1. Read your agent entry point (`AGENTS.md` or `CLAUDE.md`).
2. Read `docs/README.md`.
3. Identify the change type.
4. Read only the docs needed for that change before editing code.

## Change Routing

### Backend or API

Read:

1. `docs/architecture/overview.md`
2. `docs/architecture/module-map.md`
3. `docs/architecture/auth-and-permissions.md`
4. `docs/api/conventions.md`
5. `docs/api/route-index.md`
6. the relevant feature doc

### Frontend Page, Route, or Permission

Read:

1. `docs/architecture/module-map.md`
2. `docs/architecture/routing-and-menus.md`
3. `docs/standards/frontend-table-pages.md` when relevant
4. the relevant feature doc

### ESI, SSO, or CCP Sync

Read:

1. `docs/architecture/overview.md`
2. `docs/architecture/module-map.md`
3. `docs/architecture/runtime-and-startup.md`
4. `docs/features/current/auth-and-characters.md`
5. `docs/features/current/esi-refresh.md`
6. `docs/guides/adding-esi-feature.md`

Read local `README.md` files under `server/pkg/eve/esi/` only when the task is clearly in that area.

## Agent Rules

Do:

- treat `docs/ai/repo-rules.md` as the primary authority
- reason from committed repository artifacts and user-provided session context
- read surrounding module code, not only the file being edited
- update relevant docs when behavior, routes, runtime behavior, or standards change
- stop and reassess when blocked or looping

Do not:

- treat `docs/templates/` or local directory `README.md` files as repository-wide authority
- write future or planned behavior into current-state docs
- revert working behavior only to satisfy stale docs
- create a shadow documentation tree
- keep editing without progress

## Conflict Handling

- Use the authority order and code-vs-docs rule in `docs/ai/repo-rules.md`.
- If code and docs disagree, determine whether code drifted or docs became stale before changing either.

## Finish

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
