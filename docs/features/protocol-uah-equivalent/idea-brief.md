---
status: Draft
skill: ao-discovery
owner: "Andrii Opryshko"
reviewers: []
updated_at: "2026-11-27"
stage: "01"
depth: ""
ticket: "#65"
---

<!-- Seed of the brief: only what issue #65 and the contract-uah-equivalent PRD already state. The discovery
     interview (/ao-discovery protocol-uah-equivalent) fills the remaining sections. -->

# Idea Brief — гривневий еквівалент договору у слоті протоколу

## 1. Raw idea

Parked in the contract-uah-equivalent PRD §8: «showing the equivalent in a protocol slot (governance)».
Tracked as issue #65.

## 2. Problem

A board decision about a contract signed in EUR or USD has an amount slot in the contract's own currency. The
board and the ревізійна комісія work in hryvnias, so the protocol needs the гривневий еквівалент next to that
amount, and today it is copied by hand.

contract-uah-equivalent (#58) stores the equivalent on the contract, with its rate, дата курсу and джерело
курсу, and never recomputes it (its ADR-0002). The protocol has no way to show it yet.

## 5. Out of scope

- Computing or recomputing an equivalent in governance: the stored one is the only source.
- Currencies other than EUR and USD, rates other than the NBU official one (as in contract-uah-equivalent).

## 15. Open questions

- Is the equivalent its own слот рішення or part of the amount slot?
- Does it need a член правління's confirmation like any other slot (board-protocol AC-21), or is a value
  stored by finance confirmed by its source?
- What does the protocol show for a contract that has no equivalent yet (rate source was down at creation)?

## Related

- Issue #65; depends on #58 (`docs/features/contract-uah-equivalent/`).
- `docs/features/board-protocol/CONTEXT.md` (слот рішення), `docs/adr/` (governance and finance are separate
  contexts).
