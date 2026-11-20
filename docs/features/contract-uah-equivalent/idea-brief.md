---
status: Confirmed
owner: "Andrii Opryshko"
updated_at: "2026-11-20"
feature_size: "S"
stage: "01"
depth: easy
---

# Idea Brief — contract-uah-equivalent

Short form: the idea was picked by the owner on 2026-11-20 from three candidates for the product's first
external data source (official NBU rate · decision registry for an AI chat · counterparty lookup by ЄДРПОУ).
No competitive research and no RICE were run for a feature of this size; none is claimed.

## 1. Raw idea

> Курс НБУ для договорів: власний MCP-сервер з офіційним курсом НБУ на дату; фіча у finance — гривневий
> еквівалент договору в EUR/USD за курсом на дату підписання; ядро ходить до сервера як MCP-клієнт.

## 2. Problem

The finance context stores a contract's amount in its own currency (UAH, EUR or USD). The association's
budget, its board decisions and the ревізійна комісія's check are all in hryvnias. For a contract signed in
euros or dollars somebody looks the official rate up by hand and types the product of two numbers into a
protocol or a spreadsheet; half a year later nobody can show which rate, of which day, that number came from.

## 3. Users

The головуючий who records the contract; the ревізійна комісія who checks spending against it.

## 4. Scope

- **In:** the hryvnia equivalent of a foreign-currency contract at the official NBU rate of the day it was
  signed, stored with the contract together with the rate, its date and its source; a way to get it later
  when the rate could not be read at creation.
- **Out:** revaluation at any other day; commercial or cash rates; currencies beyond EUR and USD; a web
  screen (the finance context has no pages yet); payments and exchange differences.

## 5. Source of truth

| Number | Owner |
|---|---|
| amount, currency, signing day | the contract (entered by the board) |
| rate | the National Bank of Ukraine, read through the `nbu-rates` MCP server |
| hryvnia equivalent | Спільно: amount × rate, rounded half up to the kopiyka, computed once |

## 6. AI / deterministic boundary

Fully deterministic. The MCP server is a data source for the core API; no model takes part in the number.
The same server is also usable by an agent, which is a side benefit, not a requirement.

## 7. Risks

- The rate source is down when a contract is recorded → the contract must still be saved.
- A float sneaks into money math → rates are decimal strings end to end, arithmetic is integer.
- A rate changes after it was stored → it does not: an official rate of a past day is final, and the stored
  equivalent is never recomputed.
