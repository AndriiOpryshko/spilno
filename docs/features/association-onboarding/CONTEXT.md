---
status: Living
updated_at: "2026-11-10"
---

# Domain Context — association-onboarding

Shared terms (голова правління, член правління, реєстр протоколів, справа…) are defined in
[`../board-protocol/CONTEXT.md`](../board-protocol/CONTEXT.md); this file adds only the onboarding terms.

## Glossary

- вимкнене ОСББ — an association the оператор has disabled with a reason: its members can only read and download the signed protocols of its реєстр протоколів, nothing changes, and the оператор can enable it again. NOT a deleted association (all data is kept); NOT a member whose term ended (that is the roster, managed by the board).
- витяг з ЄДР — the state-register extract for the association's ЄДРПОУ that names its керівник (голова правління), attached to an заявка на підключення or a запит на передачу and signed with the голова's КЕП. NOT the election protocol (a decision of the association, not the register's record).
- голова правління — the elected head of the association's board, recorded in ЄДР as its керівник; the only person who may file an заявка на підключення or a запит на передачу. NOT головуючий (a per-meeting role).
- заявка на підключення — a голова правління's request to connect an association to Спільно: ЄДРПОУ, name, address, the applicant's position and term, and the КЕП-signed витяг з ЄДР; it is очікує, схвалена or відхилена with a reason, and a rejected one may be fixed and filed again. NOT a запит на передачу (for an association already connected).
- запит на передачу — a newly elected голова правління's request to take over an association already connected to Спільно, with a КЕП-signed витяг з ЄДР naming them; on approval they join the roster and the previous голова's term ends. NOT an заявка на підключення (a new association).
- КЕП — a qualified electronic signature of a person; here, the голова's signature on the витяг з ЄДР that proves the applicant is the person the register names. NOT a Google account (which proves only an e-mail address).
- оператор — the person who runs Спільно (the product owner) and decides заявки на підключення and запити на передачу and disables or enables associations; takes no decisions of any association. NOT a член правління; NOT the управляюча компанія.
- підключене ОСББ — an association whose заявка на підключення the оператор approved (or that was set up by hand before onboarding existed): its члени правління work in Спільно and it can be вимкнене ОСББ. NOT an association with an очікує заявка (nothing is created until approval).
- розбіжність імен — an advisory notice to the оператор when the applicant's typed name, the signer of the КЕП and the керівник named in the витяг з ЄДР are not the same person; it never blocks a decision by itself, but approving over it requires the оператор to name the document relied on. NOT an invalid signature (which the оператор finds while checking outside Спільно and rejects with that reason; nothing is checked at filing); NOT the extract-age warning (an extract older than 30 days, advisory and needing no document).
