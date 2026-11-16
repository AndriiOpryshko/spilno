---
id: T77
title: "Reuse InlineAlert, the generated hooks and router links across the identity screens"
layer: "ui"
deps: ["T71", "T74", "T75"]
acs: ["AC-13", "AC-21"]
files_hint:
  - "apps/web/src/features/identity/"
  - "apps/web/src/routes/"
  - "apps/web/src/components/ui/label.tsx"
  - "apps/web/src/components/ui/dialog.tsx"
  - "apps/web/src/features/governance/inline-alert.tsx"
  - "docs/design/design-system.md"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T77 — Reuse InlineAlert, the generated hooks and router links across the identity screens

## Why

[Review of 2026-11-10](../_review/review-2026-11-10.md), W5. Tasks T38–T40 and T42 named `InlineAlert` for warnings, refusals and the suspended notice; CLAUDE.md asks for generated hooks and shadcn primitives. The screens use ad-hoc alert paragraphs, hand-made query keys beside the generated ones, `<a href>` for in-app navigation, and a `label` and a `dialog` that are not the shadcn ones.

**Satisfies:** [PRD §5](../PRD.md) AC-13, AC-21 · **Blocked by:** [T71](./T71-ui-ukrainian-warnings-field-errors.md), [T74](./T74-ui-shared-problem-fallback.md), [T75](./T75-ui-mutation-errors-and-focus.md)

## What

- `InlineAlert` (moved to `components/ui` if it is to be shared across contexts) for warnings, refusals and the notice.
- `useGetSession`, `useListMemberships`, the card and storage hooks with their generated keys; invalidation through those keys.
- `<Link>` for in-app navigation; shadcn `label` and `dialog` installed, or the trimmed copies recorded in `docs/design/design-system.md`.

## Definition of Done

- [ ] Acceptance: No identity screen renders an ad-hoc alert paragraph, hand-made query key or in-app `<a href>`, and `label` and `dialog` are either the shadcn primitives or listed as deviations in the design system document.
- [ ] `grep` for `role="alert" className="text-destructive"`, `queryKey: ['` and in-app `<a href` under `features/identity` and `routes` is empty
- [ ] every existing component test and story stays green
- [ ] `mise run lint` and `mise run test` green

## Notes

Last of the UI lane on purpose: it touches the files of T66–T76.
