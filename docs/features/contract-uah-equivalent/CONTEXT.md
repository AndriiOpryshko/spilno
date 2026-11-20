---
status: Living
updated_at: "2026-11-20"
---

# Domain Context — contract-uah-equivalent

Shared terms (головуючий, ревізійна комісія, член правління…) are defined in
[`../board-protocol/CONTEXT.md`](../board-protocol/CONTEXT.md) and
[`../association-onboarding/CONTEXT.md`](../association-onboarding/CONTEXT.md).

## Glossary

- гривневий еквівалент — the amount of a foreign-currency contract in hryvnias: its amount times the офіційний курс of the day it was signed, rounded half up to the kopiyka, computed once and stored with the contract. NOT a revaluation at today's rate; NOT what the bank will actually charge on payment.
- офіційний курс — the exchange rate the National Bank of Ukraine publishes for a currency and a calendar day, hryvnias per one unit. NOT a commercial, cash or interbank rate.
- дата курсу — the calendar day the офіційний курс is for; for a contract it is always the day the contract was signed. NOT the day the rate was looked up.
- джерело курсу — who published the rate; always the NBU, read through the `nbu-rates` MCP server. NOT the MCP server itself (it relays, it does not set rates).
