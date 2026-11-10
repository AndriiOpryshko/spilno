---
id: T49
title: "Run and record the security review that PRD §6.1 requires"
layer: "docs"
deps: ["T45", "T46"]
acs: ["AC-04c", "AC-07", "AC-16b", "AC-16c", "AC-28b"]
files_hint:
  - "docs/features/association-onboarding/_review/"
owner: "Andrii Opryshko"
estimate: "M"
status: "todo"
---

# T49 — Run and record the security review that PRD §6.1 requires

## Why

[PRD](../PRD.md) §6.1 «Security review: Required» — a new sign-in path, a new role, a manual signature check, membership in several associations and the first second association.

**Satisfies:** [PRD §5](../PRD.md) AC-04c, AC-07, AC-16b, AC-16c, AC-28b · **Blocked by:** [T45](./T45-e2e-transfer-and-card.md), [T46](./T46-e2e-disable-storage-access.md)

## What

- Review the feature diff (`/security-review`): sign-in for people off the roster, the оператор's account and own filings, row-level security and the isolation test, the storage of extracts and who downloads them, the storage consent `state`, what the logs carry.
- Record `docs/features/association-onboarding/_review/security-<date>.md`: the seven abuse cases of [PRD](../PRD.md) §6.1, each with its covering test or a finding.

## Definition of Done

- [ ] Acceptance: A review record maps each of the seven abuse cases of PRD §6.1 to the test that covers it or to a finding with an owner and a date, and no High finding is open.
- [ ] every abuse case has a test name or a finding with owner and due date
- [ ] no finding of High severity is open; accepted ones are listed with who accepted them
- [ ] `mise run lint` green

## Notes

Gate before the feature is marked ready. Accepted risks already known: the own-filing guard compares accounts (SAD §11); a personal storage account with an unknown address is not detectable (AC-28b).
