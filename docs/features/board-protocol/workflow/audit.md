# Аудит контракту API: openapi.yaml vs «spilno specifics»

- **Об'єкт:** живий контракт `services/core/api/openapi.yaml` (41 операція, гілка `chore/openapi-contract-audit`).
- **Еталон:** `.claude/rules/openapi.md`, розділ «spilno specifics».
- **Режим:** read-only. Код і контракт не змінювались.

## Метод

1. Один агент зібрав список `operationId` зі спеки (41).
2. Для кожної операції окремий аудитор (sonnet) перевірив 7 правил і повернув `findings[]` з `{rule, claim, evidence: file:line}`.
   Якщо результат був зламаний або порожній, аудитора перезапускали до 2 разів.
3. Кожну знахідку незалежно перевіряв окремий агент-верифікатор, який не бачив міркувань аудитора.
   У звіті лишаються тільки підтверджені знахідки. Непідтверджені відфільтровано й пораховано.
4. Через ліміт агентів (~90) частина знахідок не дійшла до верифікатора. Вони перелічені окремо і в
   таблиці позначені `fail?(unverified)`: ні підтверджені, ні спростовані.

### Правила

| Id | Правило |
|---|---|
| R1-tag | рівно один tag, і він збігається з префіксом шляху `/<context>/` |
| R2-security | `SessionCookie`, або явне `security: []` тільки для публічних (health, sign-in, callback) |
| R3-auth-errors | захищена операція має `401 identity.session_required` і `403 identity.not_current_member` |
| R4-422-problem-json | операція з тілом має `422`; помилки як `application/problem+json` |
| R5-code-registry | кожен `code` з відповідей є в реєстрі Go (`internal/**/domain/errors.go`, `httpapi`) з тим самим рядком |
| R6-pagination-revision | списки: `?cursor=&limit=` і `{items, next_cursor}`; мутації протоколу: `protocol_revision` і `409 governance.draft_revision_conflict` |
| R7-handler | статус handler-а; для implemented: чи повертає він коди й статуси з контракту; для stub: чи не повертає недекларований статус |

## Підсумок

| Показник | Значення |
|---|---|
| Операцій усього | 41 |
| Проаудитовано | 41 |
| Підтверджених знахідок | 46 (у 23 операціях) |
| Відфільтровано (верифікатор не підтвердив) | 1 |
| Failed (аудитор не дав валідного результату після 2 перезапусків) | 0 |
| Не верифіковано через ліміт агентів | 45 знахідок у 17 операціях |
| Статус handler-ів | 7 implemented, 34 stub, 0 missing |

Головні висновки:

1. **R5: коди `identity.session_required` та `identity.not_current_member` ніде не визначені в Go.**
   Немає `internal/identity/domain/errors.go`, а `httpapi` має тільки generic `DomainError{Code}`. Рядки є лише
   в спеці (`openapi.yaml:1265`, `:1272`, приклади `:1323`, `:1325`) і в згенерованих TS-моделях. Знахідка
   наскрізна: вона стосується кожної операції, яка посилається на `SessionRequired` / `NotCurrentMember`.
   Відповідно, жоден Go-код не віддає 401/403 сесії: у route group `governance` немає session middleware
   (`governance/module.go:36-43`), а валідатор використовує `NoopAuthenticationFunc` (`httpapi/problem.go:58`).
2. **R3: шість «старих» операцій (`finance` contracts, `governance` meetings) не декларують 401/403**, хоча
   захищені глобальним `SessionCookie` (`openapi.yaml:10-11`). `signOut` декларує 401, але не 403.
3. **R7: stub-и повертають `httpapi.ErrNotImplemented`, який `StrictResponseError` перетворює на 501**
   (`httpapi/problem.go:131-134`). 501 не задекларовано в жодній операції. Це очікуваний стан для ще не
   реалізованих операцій, але контракт і runtime розходяться.
4. Не верифіковані, але варті уваги: `startPollRound`, `returnProtocolToDraft`, `markProtocolReadyForSigning`
   мають обов'язкове тіло без `422` (R4); `addAgendaItemFromCase` і `lockProtocolWithSignedCopy` змінюють
   протокол без `protocol_revision` (R6).

## Таблиця: операція × правило

Позначки: `pass`, `FAIL` (є підтверджена знахідка), `n/a`, `fail?(unconfirmed)` (аудитор сказав fail, але
верифікатор знахідку не підтвердив), `fail?(unverified)` (аудитор сказав fail, верифікації не було через ліміт).

| Операція | Handler | R1 | R2 | R3 | R4 | R5 | R6 | R7 |
|---|---|---|---|---|---|---|---|---|
| `getHealth` GET /system/health | implemented | pass | pass | n/a | n/a | n/a | n/a | pass |
| `createContract` POST /finance/contracts | implemented | pass | pass | FAIL | pass | n/a | n/a | pass |
| `listContracts` GET /finance/contracts | implemented | pass | pass | FAIL | pass | n/a | pass | pass |
| `getContract` GET /finance/contracts/{id} | implemented | pass | pass | FAIL | pass | pass | n/a | pass |
| `createMeeting` POST /governance/meetings | implemented | pass | pass | FAIL | pass | n/a | n/a | pass |
| `listMeetings` GET /governance/meetings | implemented | pass | pass | FAIL | pass | pass | pass | pass |
| `getMeeting` GET /governance/meetings/{id} | implemented | pass | pass | FAIL | pass | n/a | n/a | pass |
| `startSignIn` GET /identity/sign-in | stub | pass | pass | n/a | n/a | n/a | n/a | pass |
| `completeSignIn` GET /identity/sign-in/callback | stub | pass | pass | n/a | pass | FAIL | n/a | FAIL |
| `signOut` POST /identity/sign-out | stub | pass | pass | FAIL | n/a | FAIL | n/a | FAIL |
| `getSession` GET /identity/session | stub | pass | pass | pass | pass | FAIL | n/a | FAIL |
| `getAssociationSettings` GET /governance/settings | stub | pass | pass | pass | n/a | FAIL | n/a | FAIL |
| `saveAssociationSettings` PUT /governance/settings | stub | pass | pass | pass | pass | FAIL | n/a | FAIL |
| `listBoardMembers` GET /governance/board-members | stub | pass | pass | pass | pass | FAIL | pass | FAIL |
| `createBoardMember` POST /governance/board-members | stub | pass | pass | pass | pass | FAIL | n/a | fail?(unconfirmed) |
| `updateBoardMember` PUT /governance/board-members/{board_member_id} | stub | pass | pass | pass | pass | FAIL | n/a | FAIL ¹ |
| `listCases` GET /governance/cases | stub | pass | pass | pass | pass | FAIL | pass | n/a ² |
| `createCase` POST /governance/cases | stub | pass | pass | pass | pass | FAIL | n/a | n/a ² |
| `getCase` GET /governance/cases/{case_id} | stub | pass | pass | pass | pass | FAIL | n/a | FAIL |
| `addCaseDocument` POST /governance/cases/{case_id}/documents | stub | pass | pass | pass | pass | FAIL | n/a | pass ² |
| `downloadFile` GET /governance/files/{file_id} | stub | pass | pass | pass | n/a | FAIL | n/a | FAIL |
| `createAssistantTurn` POST /governance/meetings/{meeting_id}/assistant-turns | stub | pass | pass | pass | pass | FAIL | n/a | FAIL |
| `addAgendaItemFromCase` POST /governance/meetings/{meeting_id}/agenda-items | stub | pass | pass | pass | pass | FAIL | fail?(unverified) | fail?(unverified) |
| `getJob` GET /governance/jobs/{job_id} | stub | pass | pass | pass | pass | FAIL | n/a | n/a ² |
| `retryJob` POST /governance/jobs/{job_id}/retry | stub | pass | pass | pass | n/a | FAIL | n/a | FAIL |
| `listProtocols` GET /governance/protocols | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `getProtocol` GET /governance/protocols/{protocol_id} | stub | pass | pass | pass | n/a | fail?(unverified) | n/a | fail?(unverified) |
| `listAssistantTurns` GET /governance/protocols/{protocol_id}/assistant-turns | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `saveAgendaItemWording` PUT /governance/protocols/{protocol_id}/agenda-items/{agenda_item_id}/wording | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `addDecisionSlot` POST /governance/protocols/{protocol_id}/agenda-items/{agenda_item_id}/slots | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `confirmDecisionSlot` POST /governance/protocols/{protocol_id}/slots/{slot_id}/confirmation | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `correctDecisionSlot` POST /governance/protocols/{protocol_id}/slots/{slot_id}/values | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `startPollRound` POST /governance/protocols/{protocol_id}/poll-rounds | stub | pass | pass | pass | fail?(unverified) | fail?(unverified) | pass | fail?(unverified) |
| `castPollVote` PUT /governance/protocols/{protocol_id}/poll-rounds/{round_number}/votes | stub | pass | pass | pass | pass | fail?(unverified) | n/a ³ | fail?(unverified) |
| `returnProtocolToDraft` POST /governance/protocols/{protocol_id}/return-to-draft | stub | pass | pass | pass | fail?(unverified) | fail?(unverified) | pass | fail?(unverified) |
| `markProtocolReadyForSigning` POST /governance/protocols/{protocol_id}/ready-for-signing | stub | pass | pass | pass | fail?(unverified) | fail?(unverified) | pass | fail?(unverified) |
| `lockProtocolWithSignedCopy` POST /governance/protocols/{protocol_id}/signed-copy | stub | pass | pass | pass | pass | fail?(unverified) | fail?(unverified) | pass ² |
| `withdrawProtocol` POST /governance/protocols/{protocol_id}/withdrawal | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `copyWithdrawnProtocol` POST /governance/protocols/{protocol_id}/copy | stub | pass | pass | pass | n/a | fail?(unverified) | n/a | fail?(unverified) |
| `searchRegistry` GET /governance/registry/decisions | stub | pass | pass | pass | pass | fail?(unverified) | pass | fail?(unverified) |
| `getSignedProtocol` GET /governance/registry/protocols/{protocol_id} | stub | pass | pass | pass | pass | fail?(unverified) | n/a | fail?(unverified) |

Примітки:

1. Аудитор `updateBoardMember` поставив R7 = n/a («stub, нема з чим порівнювати»), але сам подав R7-знахідку, і
   верифікатор її підтвердив. Тому в клітинці FAIL.
2. Аудитори по-різному оцінили R7 для stub-ів. Для `listCases`, `createCase`, `getJob` вони поставили n/a, для
   `addCaseDocument`, `lockProtocolWithSignedCopy`, `startSignIn` поставили pass, а для інших stub-ів FAIL. Механізм
   скрізь однаковий: stub повертає `ErrNotImplemented`, а клієнт отримує недекларований 501. Ці клітинки варто
   читати як FAIL, але знахідок за ними не подавали й не верифікували.
3. `castPollVote`: аудитор вирішив, що це upsert голосу члена правління, а не редагування чернетки, тому
   `protocol_revision` не вимагав. Це суб'єктивна оцінка, тож знахідку він не подавав.

Для implemented-операцій у колонці R5 стоїть pass/n/a, бо вони не декларують жодного `code`. Наскрізна прогалина з
`identity.*` для них відображена в R3.

## Підтверджені знахідки

Доказ (evidence) береться від верифікатора, якщо він точніший за доказ аудитора.

### R3-auth-errors (7)

Захищена операція (глобальний `SessionCookie`, `openapi.yaml:10-11`, без override) не декларує `401
identity.session_required` та/або `403 identity.not_current_member`. Компоненти `SessionRequired` (`:1264`) і
`NotCurrentMember` (`:1270`) існують, і сусідні операції на них посилаються (наприклад, `getSession` `:234-235`).

| Операція | Твердження | Evidence |
|---|---|---|
| `createContract` | Декларує тільки 201 і 422, немає 401/403 | `services/core/api/openapi.yaml:42-57` (responses `:51-57`) |
| `listContracts` | Тільки 200 і 422, немає 401/403 | `services/core/api/openapi.yaml:57-76` (responses `:70`, `:76`) |
| `getContract` | Тільки 200, 404, 422, немає 401/403 | `services/core/api/openapi.yaml:87-94` |
| `createMeeting` | Тільки 201 і 422, немає 401/403 | `services/core/api/openapi.yaml:105-111` |
| `listMeetings` | Тільки 200 і 422, немає 401/403 | `services/core/api/openapi.yaml:113-131` (responses `:126`, `:131`) |
| `getMeeting` | Тільки 200, 404, 422, немає 401/403 | `services/core/api/openapi.yaml:142-149` |
| `signOut` | Є 401 `SessionRequired`, але немає 403 `NotCurrentMember`. ADR-0003 каже, що middleware віддає 403 для кожного захищеного маршруту | `services/core/api/openapi.yaml:204-216`; `docs/features/board-protocol/adr/0003-sign-in-with-oidc-and-postgres-sessions.md:42` |

### R5-code-registry (27)

Контракт рекламує `identity.session_required` (приклад `openapi.yaml:1323`, response `:1264-1269`) і
`identity.not_current_member` (приклад `:1325`, response `:1270-1277`). Жоден `.go` файл не містить цих рядків:
ні рукописний, ні згенерований (там тільки type alias `SessionRequired`/`NotCurrentMember`, наприклад
`identity/adapters/http/gen/identity.gen.go:89,94`, `governance/adapters/http/gen/governance.gen.go:993`). Єдиний
`domain/errors.go` належить governance (`services/core/internal/governance/domain/errors.go:17-31`, тільки
`governance.*`). `internal/identity` не має `domain/`. `httpapi/problem.go:145-153` визначає лише generic
`DomainError`.

| Операція | Код(и) | Evidence (посилання операції → визначення коду) |
|---|---|---|
| `completeSignIn` | `identity.session_required` | `openapi.yaml:197-202` → `:1323` |
| `completeSignIn` | `identity.not_current_member` | `openapi.yaml:204` → `:1325` |
| `signOut` | `identity.session_required` | `openapi.yaml:216` → `:1265-1269`, `:1323` |
| `getSession` | обидва | `openapi.yaml:234-235` → `:1264-1277` |
| `getAssociationSettings` | обидва | `openapi.yaml:250-251` → `:1323`, `:1325` |
| `saveAssociationSettings` | обидва | `openapi.yaml:278-279` → `:1264-1277` |
| `listBoardMembers` | обидва | `openapi.yaml:314-315` → `:1323`, `:1325` |
| `createBoardMember` | обидва | `openapi.yaml:338-339` → `:1323`, `:1325` |
| `createBoardMember` | `identity.not_current_member` | `openapi.yaml:339` → `:1325` |
| `updateBoardMember` | `identity.session_required` | `openapi.yaml:370` → `:1265` |
| `updateBoardMember` | `identity.not_current_member` | `openapi.yaml:371` → `:1270-1272` |
| `listCases` | обидва | `openapi.yaml:398-399` → `:1323`, `:1325` |
| `listCases` | `identity.not_current_member` | `openapi.yaml:399` → `:1270-1277`, `:1325` |
| `createCase` | `identity.session_required` | `openapi.yaml:418` → `:1265`, `:1323` |
| `createCase` | `identity.not_current_member` | `openapi.yaml:419` → `:1277`, `:1325` |
| `getCase` | обидва | `openapi.yaml:440-441` → `:1264-1277` |
| `getCase` | `identity.not_current_member` | `openapi.yaml:441` → `:1270-1272` |
| `addCaseDocument` | обидва | `openapi.yaml:472-473` → `:1322-1324` |
| `downloadFile` | `identity.session_required` | `openapi.yaml:508` → `:1323` |
| `downloadFile` | `identity.not_current_member` | `openapi.yaml:509` → `:1325` |
| `createAssistantTurn` | обидва | `openapi.yaml:549-550` → `:1264-1272` |
| `createAssistantTurn` | `identity.not_current_member` | `openapi.yaml:1270-1277`, `:1325` |
| `addAgendaItemFromCase` | обидва | `openapi.yaml:589-590` → `:1264-1272` |
| `getJob` | `identity.session_required` | `openapi.yaml:615` → `:1265`, `:1323` |
| `getJob` | `identity.not_current_member` | `openapi.yaml:616` → `:1270-1273` |
| `retryJob` | обидва | `openapi.yaml:632-633` → `:1323`, `:1325` |
| `retryJob` | `identity.not_current_member` | `openapi.yaml:1325` |

Деякі аудитори подали окремою знахідкою «обидва коди» і ще окремою `not_current_member`, тому в 10 операціях по дві
R5-знахідки. Суть одна: у реєстрі Go бракує двох констант.

### R7-handler (12)

Stub повертає `httpapi.ErrNotImplemented`. `StrictResponseError` (`services/core/internal/platform/httpapi/problem.go:131-134`,
підключено як `ResponseErrorHandlerFunc` у `governance/module.go:40` та `identity/module.go:37`) відповідає на нього
501 problem+json без `code`. 501 не задекларовано ніде в контракті, а задекларовані відповіді stub не повертає ніколи.
Верифікатори зробили одне уточнення: спільний `RequestValidator` (`problem.go:50-81`) ще до handler-а може віддати
`422` без `code`, але це відповідь middleware, а не handler-а.

| Операція | Твердження | Evidence |
|---|---|---|
| `completeSignIn` | 501 замість задекларованих 302/401/403 | `services/core/internal/identity/adapters/http/handler.go:26-28`; `openapi.yaml:186-203` |
| `signOut` | Не видаляє сесію, не чистить cookie, не повертає 204/401 | `services/core/internal/identity/adapters/http/handler.go:30-31` |
| `getSession` | 501 замість 200/401/403 | `services/core/internal/identity/adapters/http/handler.go:18-19`; `openapi.yaml:217-235` |
| `getAssociationSettings` | 501 замість 200/401/403/409 `governance.settings_missing` | `services/core/internal/governance/adapters/http/stubs.go:125-127`; `openapi.yaml:239-257` |
| `saveAssociationSettings` | 501 замість 200/401/403/409/422 | `services/core/internal/governance/adapters/http/stubs.go:129-131`; `openapi.yaml:258-290` |
| `listBoardMembers` | 501 замість 200/401/403/422 | `services/core/internal/governance/adapters/http/stubs.go:13-14`; `openapi.yaml:306`, `:314-316` |
| `updateBoardMember` | 501 замість 200/401/403/404/422 `governance.roster_entry_invalid` | `services/core/internal/governance/adapters/http/stubs.go:21-23`; `openapi.yaml:348-378` |
| `getCase` | 501 замість 200/401/403/404 | `services/core/internal/governance/adapters/http/stubs.go:33-34`; `openapi.yaml:425-443` |
| `getCase` | 501 пише спільний обробник, і його немає серед responses | `services/core/internal/platform/httpapi/problem.go:132-133`; `openapi.yaml:428-442` |
| `downloadFile` | 501 замість 200 (файл)/401/403/404/503 | `services/core/internal/governance/adapters/http/stubs.go:41-43`; `openapi.yaml:493-511` |
| `createAssistantTurn` | 501 замість 202 і кодів 409/422/503 | `services/core/internal/governance/adapters/http/stubs.go:57-58`; `openapi.yaml:514-568` |
| `retryJob` | Не повертає 202/404/409 `governance.transition_not_allowed`. Константа `CodeTransitionDenied` є, але не використовується | `services/core/internal/governance/adapters/http/stubs.go:49-51`; `services/core/internal/governance/domain/errors.go:27` |

### Звірка лічильника

Підтверджених знахідок 46: R3 = 7, R5 = 27, R7 = 12.

## Відфільтровані, failed, неверифіковані

### Відфільтровано верифікатором: 1

Одна знахідка не пройшла незалежну перевірку. За таблицею це R7 у `createBoardMember`: аудитор поставив fail
(stub → 501), підтвердженої знахідки немає, клітинка `fail?(unconfirmed)`. Текст відкинутої знахідки workflow не
зберігає.

### Failed-операції: 0

Усі 41 аудитори повернули валідний результат.

### Не верифіковано через ліміт агентів: 45 знахідок, 17 операцій

| Операція | Правило | Твердження аудитора | Evidence |
|---|---|---|---|
| `addAgendaItemFromCase` | R6 | Мутація чернетки протоколу без `protocol_revision` (вхід тільки `case_id`), без 409 `governance.draft_revision_conflict` | `services/core/api/openapi.yaml:1920` |
| `addAgendaItemFromCase` | R6 | 409 перелічує тільки `settings_missing` і `protocol_locked`, а `draft_revision_conflict` пропущено | `services/core/api/openapi.yaml:590` |
| `addAgendaItemFromCase` | R7 | Stub → 501 | `services/core/internal/governance/adapters/http/stubs.go:53` |
| `listProtocols` | R5 | `identity.session_required` / `identity.not_current_member` не в реєстрі Go | `services/core/api/openapi.yaml:1265` |
| `listProtocols` | R5 | `identity.not_current_member` не в реєстрі Go | `services/core/api/openapi.yaml:1272` |
| `listProtocols` | R7 | Stub, 200/401/403/422 не повертаються | `services/core/internal/governance/adapters/http/stubs.go:61` |
| `getProtocol` | R5 | identity-коди не в реєстрі Go | `services/core/api/openapi.yaml:1323` |
| `getProtocol` | R7 | Stub → 501 | `services/core/internal/governance/adapters/http/stubs.go:65` |
| `saveAgendaItemWording` | R5 | identity-коди не в реєстрі Go | `services/core/api/openapi.yaml:1265` |
| `saveAgendaItemWording` | R5 | `identity.not_current_member` не в реєстрі Go | `services/core/api/openapi.yaml:1272` |
| `saveAgendaItemWording` | R7 | Stub, 200/401/403/404/409/422 не повертаються | `services/core/internal/governance/adapters/http/stubs.go:73` |
| `listAssistantTurns` | R5 | `identity.session_required` не в реєстрі Go | `services/core/api/openapi.yaml:1323` |
| `listAssistantTurns` | R5 | `identity.not_current_member` не в реєстрі Go | `services/core/api/openapi.yaml:1325` |
| `listAssistantTurns` | R7 | Stub → 501 | `services/core/internal/governance/adapters/http/stubs.go:77-79` |
| `startPollRound` | R4 | Обов'язкове тіло, але 422 не задекларовано (тільки 201/401/403/404/409) | `services/core/api/openapi.yaml:859` |
| `startPollRound` | R5 | identity-коди не в реєстрі Go | `services/core/api/openapi.yaml:1265` |
| `startPollRound` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:85` |
| `addDecisionSlot` | R5 | identity-коди не в реєстрі Go | `services/core/api/openapi.yaml:1322` |
| `addDecisionSlot` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:69` |
| `confirmDecisionSlot` | R5 | identity-коди не в реєстрі Go | `services/core/api/openapi.yaml:1323` |
| `confirmDecisionSlot` | R5 | `ErrorRegistry` governance містить тільки `governance.*` | `services/core/internal/governance/domain/errors.go:68` |
| `confirmDecisionSlot` | R7 | Stub → 501 | `services/core/internal/governance/adapters/http/stubs.go:105` |
| `correctDecisionSlot` | R5 | identity-коди не в реєстрі Go | `services/core/api/openapi.yaml:1264` |
| `correctDecisionSlot` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:109` |
| `markProtocolReadyForSigning` | R4 | Обов'язкове тіло `ProtocolRevisionInput`, 422 немає | `services/core/api/openapi.yaml:945` |
| `markProtocolReadyForSigning` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:962` |
| `markProtocolReadyForSigning` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:93` |
| `returnProtocolToDraft` | R4 | Обов'язкове тіло, 422 немає | `services/core/api/openapi.yaml:930` |
| `returnProtocolToDraft` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:1323` |
| `returnProtocolToDraft` | R7 | Stub → 501 | `services/core/internal/governance/adapters/http/stubs.go:97` |
| `withdrawProtocol` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:1323` |
| `withdrawProtocol` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:113-115` |
| `castPollVote` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:1265` |
| `castPollVote` | R5 | `identity.not_current_member` не в Go | `services/core/api/openapi.yaml:1272` |
| `castPollVote` | R7 | Stub, коди `vote_on_behalf`/`poll_not_open`/... не реалізовані | `services/core/internal/governance/adapters/http/stubs.go:89` |
| `copyWithdrawnProtocol` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:1322` |
| `copyWithdrawnProtocol` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:81` |
| `lockProtocolWithSignedCopy` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:1323` |
| `lockProtocolWithSignedCopy` | R6 | `SignedCopyUpload` без `protocol_revision`, хоча 409 `draft_revision_conflict` задекларовано (його прив'язано до розбіжності `print_copy_id`) | `services/core/api/openapi.yaml:2264` |
| `searchRegistry` | R5 | identity-коди не в Go | `services/core/api/openapi.yaml:1264` |
| `searchRegistry` | R5 | `identity.not_current_member` не в Go | `services/core/api/openapi.yaml:1270` |
| `searchRegistry` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:117` |
| `getSignedProtocol` | R5 | `identity.session_required` не в Go | `services/core/api/openapi.yaml:1323` |
| `getSignedProtocol` | R5 | `identity.not_current_member` не в Go | `services/core/api/openapi.yaml:1325` |
| `getSignedProtocol` | R7 | Stub | `services/core/internal/governance/adapters/http/stubs.go:121-123` |

R5- і R7-знахідки з цього списку за змістом збігаються з уже підтвердженими в інших операціях (той самий механізм,
ті самі файли), тож дуже ймовірно, що вони теж правдиві. Нові за змістом і варті окремої перевірки тільки R4
(`startPollRound`, `returnProtocolToDraft`, `markProtocolReadyForSigning`) та R6 (`addAgendaItemFromCase`,
`lockProtocolWithSignedCopy`).
