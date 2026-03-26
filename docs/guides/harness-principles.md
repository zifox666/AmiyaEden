---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-26
source_of_truth:
  - docs/ai/repo-rules.md
  - docs/README.md
---

# Harness Engineering Principles

## Purpose

This guide explains how the repository is structured to support reliable AI-assisted engineering work.

It is not the rule source. Repository-wide rules live in `docs/ai/repo-rules.md` and the standards under `docs/standards/`.

Use this document to understand why the harness exists, how its parts fit together, and how to evolve it without weakening repository discipline.

This is primarily a maintainer reference, not an active-work doc for coding agents.

## What Harness Engineering Means

Harness engineering is the design of the execution environment around an AI coding agent.

It is distinct from:

- prompt engineering: what instructions are given
- context engineering: what information is made visible
- harness engineering: how the full working system is structured, constrained, and verified

Conceptually:

- prompt engineering -> what should be asked
- context engineering -> what should be shown
- harness engineering -> how the full environment should control quality, feedback, and drift

Harness engineering includes:

- repository rules
- documentation layout
- verification requirements
- change-completion criteria
- feedback loops
- drift prevention

## Why It Matters in AmiyaEden

This repository is designed for human-agent collaboration in a codebase with:

- backend and frontend contract coupling
- layered architecture
- permission boundaries
- localization requirements
- long-lived documentation that must stay aligned with implementation

Without a harness, agents tend to:

- introduce local patterns instead of following repository conventions
- change one side of a contract without updating the other
- treat passing builds as sufficient verification
- drift away from architectural boundaries
- leave documentation stale after behavior changes

The harness reduces those failure modes by making repository rules explicit and by requiring verification before work is considered complete.

## Core Model

A reliable harness in this repository has three parts:

1. context availability
2. architectural constraints
3. feedback and entropy control

## 1. Context Availability

Agents can only reason from information that is visible in the repository or explicitly provided in the session.

In this repository, important context is committed into:

- `docs/ai/repo-rules.md` for repository-wide agent rules (loaded via `AGENTS.md` and `CLAUDE.md`)
- `docs/architecture/` for current system structure
- `docs/features/current/` for current feature behavior and invariants
- `docs/api/` for route and contract information
- `docs/standards/` for enforceable engineering rules
- `docs/guides/` for procedural guidance
- document front matter for status, ownership, and source-of-truth metadata

Information that exists only in chat history, external documents, or team memory is not reliable repository context. If it matters to implementation, it must be committed into the repository.

### Practical Rule

If an agent repeatedly misses an important fact, the problem is usually not the agent. The repository is probably missing a visible, maintained source for that fact.

## 2. Architectural Constraints

Constraints make the valid solution space smaller. That improves reliability.

In AmiyaEden, the harness constrains work through:

- dependency direction rules
- layer responsibility rules
- contract synchronization requirements
- localization requirements
- type-safety rules
- repository-specific frontend and backend patterns

Examples:

- business logic belongs in services, not handlers, repositories, or Vue views
- backend and frontend contract changes must stay synchronized
- user-facing strings must be localized
- repository patterns should be reused before new abstractions are introduced

These constraints matter because agents are good at producing plausible code. The harness must ensure that plausible code is also repository-correct code.

## 3. Feedback and Entropy Control

A harness is incomplete without feedback loops.

Agents need signals that tell them whether a change is:

- valid
- structurally correct
- behaviorally safe
- complete

This repository uses three kinds of feedback.

### Build-Level Feedback

Build-level feedback answers whether the code still passes basic technical checks.

Common commands:

See `docs/standards/testing-and-verification.md` for the full command list.

These checks are necessary, but not sufficient.

### Structural Feedback

Structural feedback answers whether the change still follows repository architecture and conventions.

Examples:

- dependency-layering rules
- anti-pattern checklists
- completion checklists
- repository-specific frontend and backend standards

Structural feedback prevents changes that compile but violate the intended architecture.

### Behavioral Feedback

Behavioral feedback answers whether the system still does what it is supposed to do.

Examples:

- regression tests
- contract tests
- feature invariants documented in feature docs
- permission boundaries documented in architecture and API docs

Behavioral feedback is especially important for bug fixes, contract changes, fallback logic, filtering logic, and permission-sensitive flows.

## Feedback Hierarchy

Not all signals provide the same confidence.

From weakest to strongest:

1. build passes
2. lint and type checks pass
3. structure still matches repository rules
4. behavior is covered by regression or behavior-level tests
5. docs and implementation remain synchronized

A completed change should satisfy the strongest level that reasonably applies.

## Common Failure Modes

The harness exists to reduce recurring agent failure modes.

| failure mode | symptom | typical correction |
| --- | --- | --- |
| over-abstraction | one-off helpers or utilities introduced without reuse | prefer existing patterns and keep changes scoped |
| under-testing | build success treated as sufficient verification | apply testing and completion standards |
| documentation drift | behavior changed but docs were left stale | update docs in the same change |
| layer violation | business logic placed in handlers, repositories, or views | re-check dependency and responsibility standards |
| contract drift | backend and frontend fall out of sync | apply the contract synchronization rules |
| scope creep | unrelated cleanup mixed into the requested task | re-check original scope and remove extra changes |
| loop behavior | repeated edits without progress | re-read standards, inspect current code, change approach |

## Harness Components in This Repository

The harness is distributed across several document types.

### Rule Sources

Use these for repository-wide constraints:

- `docs/ai/repo-rules.md`
- `docs/standards/*.md`

### Current-State Sources

Use these to understand how the repository currently works:

- `docs/architecture/*.md`
- `docs/features/current/*.md`
- `docs/api/*.md`

### Procedural Sources

Use these for implementation and verification workflow:

- `docs/guides/*.md`

### Metadata Layer

Front matter provides operational metadata such as:

- status
- document type
- owner
- last reviewed date
- source of truth

That metadata helps maintain clarity about what is current, authoritative, and still maintained.

## How to Improve the Harness

When a recurring agent mistake appears, improve the harness rather than only correcting the individual change.

Preferred responses:

- if agents miss repository rules, strengthen the relevant standard
- if agents miss feature behavior, add or improve the feature doc
- if agents violate architecture, tighten layering or completion guidance
- if agents leave stale docs behind, strengthen documentation governance or completion checks
- if agents repeatedly make the same avoidable mistake, add a clearer repository-visible rule or checklist entry

Do not respond by creating duplicate documentation trees, shadow rules, or parallel standards.

## How to Use This Document

### For engineers

Use this guide to understand why the repository is documented and constrained the way it is.

### For agents

This document is explanatory, not primary. Apply repository rules from:

- `docs/ai/repo-rules.md`
- `docs/standards/*.md`

Use this guide only when you need the reasoning behind those rules.

### For harness maintainers

When changing repository-wide agent rules:

1. update the relevant rule source
2. update the related standard if procedural detail changed
3. update this guide if the harness model or rationale changed

## Related Documents

- `docs/ai/repo-rules.md`
- `docs/README.md`
- `docs/standards/dependency-layering.md`
- `docs/standards/testing-and-verification.md`
- `docs/standards/pre-completion-checklist.md`
- `docs/standards/documentation-governance.md`
- `docs/guides/testing-guide.md`
- `docs/guides/regression-test-plan.md`
