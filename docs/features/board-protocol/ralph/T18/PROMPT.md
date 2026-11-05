# PROMPT.md · Ralph loop (T18 — poll rounds and the own-vote rule)

## Контекст

Ти в репозиторії spilno (Go 1.27 модульний моноліт `services/core` + React). Кожна ітерація
починається з нуля: стан між ітераціями живе лише на диску — у git, у файлах і в
`docs/features/board-protocol/ralph/T18/progress.md` (прочитай його першим, якщо він є).

Прочитай перед роботою:
- `CLAUDE.md` — конвенції репо, особливо «Layout rules» і «Git».
- `docs/features/board-protocol/tasks/T18-poll-domain.md` — story (AC-30, AC-31, AC-05e, AC-32, AC-41, AC-05d).
- `docs/features/board-protocol/PRD.md`, розділи `### AC-30`, `### AC-31`, `### AC-41`, `### AC-05d`, `### AC-05e`, `### AC-32` — тексти AC.
- `services/core/internal/governance/domain/poll_test.go` — приймальний тест; його шапка «Expected API» — контракт, який треба реалізувати.
- `services/core/internal/governance/domain/protocol.go` і `errors.go` — референс стилю: `Protocol`, `EnsurePollOpen`, `Effects.ClosePollRound`, `DomainError`, `ErrorRegistry`.

Конвенції, що важливі тут: `domain/` — лише stdlib (depguard); помилки — `&DomainError{Code: ...}`
з константою `Code...` в `errors.go` (код `governance.vote_on_behalf` уже є в `ErrorRegistry`
літералом — заведи для нього константу і використай її там); doc-коментар на кожному
експортованому імені, як у `protocol.go`.

## Завдання

Створи `services/core/internal/governance/domain/poll.go` з API із шапки `poll_test.go`, щоб цей тест став зеленим, і закрий T18 одним комітом.

## Критерій завершення

Бінарний: `docs/features/board-protocol/ralph/T18/check.sh` друкує `PASS` і виходить з кодом 0.
Він перевіряє рівно це:
1. `poll.go` існує, `poll_test.go` байт-у-байт той самий (sha256).
2. `go -C services/core test ./internal/governance/...` зелений.
3. `go -C services/core tool -modfile=tools/go.mod golangci-lint run ./internal/governance/...` без зауважень.
4. Рядок T18 у `docs/features/board-protocol/tasks/tracker.md` має статус `done` (інші рядки не чіпай).
5. На гілці є коміт `feat(governance): <imperative summary>`, який містить `poll.go`, `poll_test.go`,
   `tracker.md` (і за потреби `errors.go`) — і нічого іншого; незакомічених змін у `services/core` немає.
   Тіло коміту: чому, потім трейлери `SDLC-Task: T18`, по рядку `SDLC-AC: <id>` на кожен AC story,
   `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` (формат — як у `git show 5c6aba9`).

Порядок ітерації: прочитай `progress.md` → зроби наступний крок → запусти `check.sh` →
допиши в `progress.md` один рядок (що зробив, що показав `check.sh`).
Вихід: щойно `check.sh` надрукував `PASS` — виконай `touch docs/features/board-protocol/ralph/T18/DONE` і завершуй.
Якщо та сама причина FAIL повторюється третю ітерацію поспіль і ти не можеш її усунути, не міняючи тест, —
запиши причину в `docs/features/board-protocol/ralph/T18/BLOCKED` і завершуй.

## Не роби

- Не змінюй `poll_test.go` та інші тести; червоний тест — виправляй код.
- Не змінюй `CLAUDE.md`, story-файли, PRD, `check.sh`, `PROMPT.md`.
- Не стейдж нічого, крім файлів із пункту 5 (`git add <path>`, ніколи `git add -A` / `git add .`);
  файли в `docs/features/board-protocol/ralph/` не комітиш.
- Не роби `git push`, не відкривай PR, не переписуй історію (`--amend` теж ні — помилка → новий коміт).
- Не додавай нових залежностей і нічого поза `governance/domain` і трекером.
