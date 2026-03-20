---
status: active
doc_type: guide
owner: engineering
last_reviewed: 2026-03-20
---

# Reference Assets

## Purpose

This folder stores large offline reference files that can help with development, investigation, or one-off data work, but are not the source of truth for runtime behavior.

## Files

- `esi-openapi.json`
  - Historical ESI OpenAPI snapshot restored from repo history
  - Useful when exploring CCP route shapes offline
  - Do not treat it as the current runtime contract without re-checking CCP docs

- `sde-schema.sql`
  - Historical SDE SQL dump/schema reference restored from repo history
  - Useful for understanding legacy table layouts and import expectations
  - Do not treat it as the current live database schema for this app

## Usage Rules

- Implementation truth still lives in code under `server/` and in the active docs under `docs/`
- If these assets are refreshed later, note the source and refresh date in this file
- Avoid writing product rules that depend solely on these snapshots
