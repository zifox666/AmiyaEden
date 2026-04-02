---
status: active
doc_type: standard
owner: engineering
last_reviewed: 2026-04-03
source_of_truth:
  - static/src/utils/common/isk.ts
  - static/src/utils/common/index.ts
---

# ISK Formatting Standard

## Scope

Applies to all user-facing ISK-denominated values in the frontend UI.

Does not apply to Fuxi Coin or any other non-ISK currency display.

## Shared Helper

- Import ISK helpers from `@/utils/common`.
- Use `formatIskPlain(value)` for exact ledger-style amounts.
- Use `formatIskSmart(value)` for compact summary amounts.
- Use `iskToMillionInput(value)` and `millionInputToIsk(value)` only for editors that intentionally expose million-based numeric inputs while storing ISK.
- Generic number-rendering components may render an exact numeric ISK value from raw number props when they do not introduce separate ISK string-formatting logic; any supporting ISK summary text must still use the shared helper.

## Approved Display Styles

### Plain ISK Value Style

- Output the full numeric value with `,` grouping and exactly `2` decimals.
- Do not abbreviate units.
- Example: `711,103,702.38`

### Smart Abbreviation Style

- Output exactly `2` decimals plus an uppercase unit suffix.
- Insert one space before the suffix.
- Allowed suffixes are `K`, `M`, `B`, and `T`.
- Values that round to `1000.00` in one unit must promote to the next unit.
- Example: `711.10 M`

## Surface Rules

- Use `formatIskPlain` for wallet balances, wallet journals, NPC kill report amounts, and newbro ISK-denominated displays.
- Use `formatIskSmart` for SRP displays, contract list/detail values, and compact ISK summaries such as the dashboard console wallet description.
- When one surface intentionally shows both styles, keep the exact numeric value plain and the supporting summary text smart.

## Prohibited Patterns

- Do not define local `formatISK` helpers in views, hooks, or components.
- Do not format ISK with inline `Intl.NumberFormat('en-US', ...)` or `toLocaleString('en-US', ...)` in UI renderers.
- Do not build ad-hoc `K` / `M` / `B` / `T` strings with `toFixed()`.

## Checklist

- [ ] All explicit ISK string-formatting logic routes through `static/src/utils/common/isk.ts`
- [ ] Exact-value surfaces use `formatIskPlain`
- [ ] Compact summary surfaces use `formatIskSmart`
- [ ] Million-based editors use only `iskToMillionInput` and `millionInputToIsk`
- [ ] Fuxi Coin formatting remains independent from the ISK helpers