---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-03-24
source_of_truth:
  - docs/README.md
  - docs/ai/repo-rules.md
---

# Documentation Governance Standard

## Scope

This standard governs canonical repository documentation under the repository root and `docs/`. This includes the agent entry points (`AGENTS.md`, `CLAUDE.md`) which delegate to `docs/ai/repo-rules.md`.

## Core Rules

- Each document must have a single primary responsibility.
- Each class of fact must have a single canonical source.
- Current implementation, engineering rules, and future proposals must be stored separately.
- Do not maintain a second parallel documentation tree for the same subject.
- Repository-level canonical documentation belongs only in `docs/` and the agent entry points (`AGENTS.md`, `CLAUDE.md`) which delegate to `docs/ai/repo-rules.md`.
- The root `README.md` may serve as an onboarding or product-facing entry point, but it does not define engineering rules. If conflicts exist, `docs/ai/repo-rules.md` and `docs/` take precedence.
- Subdirectory `README.md` files are local implementation notes only. They must not redefine repository-wide rules, route surfaces, or product behavior.

## Document Types

| doc_type | directory | purpose |
| --- | --- | --- |
| `agent-rules` | `docs/ai/` | shared agent rule source included by `AGENTS.md` and `CLAUDE.md` |
| `standard` | `docs/standards/` | required rules, prohibitions, and recommended practices |
| `architecture` | `docs/architecture/` | how the current system works |
| `api` | `docs/api/` | routes, authentication, and response conventions |
| `feature` | `docs/features/current/` | current module behavior, entry points, permissions, and invariants |
| `guide` | `docs/guides/` | step-by-step operating instructions |
| `reference` | `docs/reference/` | offline reference assets; not authoritative for current implementation |
| `draft` | `docs/specs/draft/` | proposals, enhancements, and unimplemented designs |
| `template` | `docs/templates/` | templates for creating new documents |

## Front Matter Requirements

All new canonical documents must include YAML front matter with at least the following fields:

- `status`
- `doc_type`
- `owner`
- `last_reviewed`
- `source_of_truth`

Example front matter:

```yaml
status: active  
doc_type: feature  
owner: engineering  
last_reviewed: 2026-03-24  
source_of_truth:  
  - server/internal/router/router.go
```

Recommended fields:

- `source_of_truth`
- `supersedes`
- `related_docs`

Template rules:

- files under `docs/templates/*` must use `status: template`
- templates must state clearly that they are templates and do not describe the current implementation

## File Naming

- Use `kebab-case`
- Name files by scope, not by temporary conclusions
- Do not use names that will age quickly, such as `new-`, `final-`, `latest-`, or `v2-`

Preferred examples:

- `auth-and-permissions.md`
- `runtime-and-startup.md`
- `route-index.md`

## Minimum Structure by Document Type

### standard

- scope
- core rules
- allowed exceptions
- checklist

### architecture

- scope
- current implementation
- key entry files
- invariants

### api

- base URL, authentication, and response conventions
- route index or interface list
- explicit permission boundaries where relevant
- synchronization requirements for changes

### feature

- module purpose
- current entry points
- permission boundaries
- key invariants
- primary code files

### reference

- asset purpose
- file list
- non-authoritative status
- usage limits or refresh guidance

### draft

- background
- current status
- proposal
- open questions
- explicit statement that it is not yet implemented

## When to Create a New Document

Create a new document when:

- a new feature module is large enough to stand on its own
- a new standard will be reused across multiple modules
- a proposal is not yet implemented but needs ongoing discussion

Do not create a new document when:

- it only repeats an existing route table from another angle
- it only rewrites an existing rule
- it only records a temporary discussion outcome
- it creates a subdirectory `README.md` that duplicates canonical documentation already maintained in `docs/`

## Update Rules

- Behavior changes and documentation updates must be made in the same change.
- When changing document status or scope, update `last_reviewed`.
- When a document moves from `draft` to active canonical status, move it to the correct directory instead of only renaming the title.
- When deleting or merging documents, remove stale references so no shadow entry points remain.

## Canonical Sources

Certain facts have a designated single source. Do not redefine or duplicate these in other documents; reference them instead.

| fact | canonical source |
| --- | --- |
| verification commands (lint, test, build) | `docs/standards/testing-and-verification.md § Default Commands` |

When adding a new category of facts that appears in multiple documents, designate one canonical source here and convert all other occurrences to references.

## Anti-Patterns

Avoid the following:

- duplicating the same role list or rules across README files, guides, and feature docs
- redefining verification commands outside `docs/standards/testing-and-verification.md § Default Commands`
- turning the root `README.md` into a competing engineering standard beside `docs/ai/repo-rules.md` and `docs/`
- mixing future plans into current-state documents
- maintaining a second parallel documentation tree that conflicts with canonical docs
- citing code too vaguely for readers to locate the real entry files
