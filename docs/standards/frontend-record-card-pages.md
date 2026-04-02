---
status: active
doc_type: standard
owner: frontend
last_reviewed: 2026-04-02
source_of_truth:
  - static/src/assets/styles/core/app.scss
  - static/src/views/newbro/select-mentor/index.vue
  - static/src/views/system/mentor-reward-stages/index.vue
---

# Frontend Record Card Page Standard

## Scope

This standard applies to frontend pages whose primary content is a repeated record list rendered as cards, stacked editors, or unpaginated editor tables embedded inside cards.

Use this for pages such as candidate directories, approval card lists, reward-stage editors, and other layouts where the record count can grow beyond a small fixed number.

Use `docs/standards/frontend-table-pages.md` instead when the primary list is a paginated page-level table. This standard covers the overflow behavior of repeated records rendered inside cards, including editable row tables that are not using the table-page pattern.

## Core Rules

- Every repeated-record section must have a declared growth strategy: either page expansion or internal overflow.
- Implicit clipping is forbidden. User data must not become unreachable because a page root, card body, tab pane, or wrapper hides overflow without providing a scroll owner.
- Default to page expansion for unbounded card lists and staged editors.
  - Let the page height grow naturally.
  - Do not use `art-full-height` on pages whose main content is an unbounded record card list unless a descendant explicitly owns scrolling.
  - Prefer the application shell scroll path when the list should simply grow with content.
- Internal scrolling is allowed only when the scroll owner is explicit.
  - The scrolling element must use `overflow: auto`, `overflow-y: auto`, or `ElScrollbar`.
  - If the page uses `art-full-height`, complete the height chain with `display: flex`, `flex-direction: column`, and `min-height: 0` through each intermediate wrapper that participates in layout.
- A card that can contain many records must either expand with the page or contain an explicit inner scroll region. It must never rely on hidden overflow as the only constraint.
- Mixed pages may combine this standard with `docs/standards/frontend-table-pages.md`, but each section must have one clear overflow owner.

## Allowed Exceptions

- Fixed-height dashboards or metric cards whose content is intentionally bounded.
- Dialog or drawer content where the dialog container already owns scrolling.
- Static summary cards with a guaranteed small item count that cannot grow with live data or configuration.

## Checklist

- If the record count doubles, does the content still remain reachable?
- If the page uses `art-full-height`, is there a single explicit scroll owner for every unbounded section?
- If the page does not need internal scrolling, did you avoid `art-full-height` and let the page expand instead?
- Did you avoid adding `overflow: hidden` on a record container without a nested scroll region?
- If the page mixes cards and tables, did you apply the card-list rule to cards and the table rule to tables separately?