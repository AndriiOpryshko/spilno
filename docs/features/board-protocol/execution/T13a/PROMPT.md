# PROMPT.md · Ralph loop (T13a — the Files port and its in-memory fake)

## Контекст

Ти в репозиторії spilno (Go модульний моноліт `services/core` + React). Кожна ітерація
починається з нуля: стан між ітераціями живе лише на диску — у git, у файлах і в
`docs/features/board-protocol/execution/T13a/progress.md` (прочитай його першим, якщо він є).

Прочитай перед роботою:
- `CLAUDE.md` — конвенції репо, особливо «Layout rules» і «Git».
- `docs/features/board-protocol/tasks/T13a-files-port-fake.md` — story і Definition of Done.
- `services/core/internal/governance/adapters/filestest/contract.go` і
  `services/core/internal/governance/adapters/filesfake/fake_test.go` — приймальні тести;
  їхні шапки «Expected API» — контракт, який треба реалізувати.
- `services/core/internal/governance/ports/ports.go` і `domain/errors.go` — референс стилю.

## Завдання

Створи `services/core/internal/governance/ports/files.go` (порт `Files`) і
`services/core/internal/governance/adapters/filesfake/fake.go` (in-memory fake) з API із цих шапок,
додай `CodeStorageUnavailable` в `domain/errors.go` і використай його в `ErrorRegistry`, щоб тести стали
зеленими, і закрий T13a одним комітом.

## Критерій завершення

Бінарний: `docs/features/board-protocol/execution/T13a/check.sh` друкує `PASS` і виходить з кодом 0.
Запускай саме так, цим шляхом, без `bash` попереду. Він перевіряє:
1. обидва файли існують, приймальні тести байт-у-байт ті самі (sha256);
2. `go -C services/core test -count=1 ./internal/governance/...` зелений;
3. `go -C services/core tool -modfile=tools/go.mod golangci-lint run ./internal/governance/...` без зауважень;
4. `ports` і `filesfake` імпортують лише stdlib і пакети цього модуля;
5. рядок T13a у `docs/features/board-protocol/tasks/tracker.md` має статус `done` (інші рядки не чіпай);
6. на гілці є коміт `feat(governance): <imperative summary>` з `ports/files.go`, `filesfake/fake.go`,
   `tracker.md` (і за потреби `domain/errors.go`) — і нічого іншого; незакомічених і невідстежуваних
   файлів у `services/core` немає. Тіло коміту: чому, потім трейлери `SDLC-Task: T13a`, по рядку
   `SDLC-AC: <id>` на кожен AC story, `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`.

Порядок ітерації: прочитай `progress.md` → зроби наступний крок → запусти `check.sh` →
допиши в `progress.md` один рядок (що зробив, що показав `check.sh`).
Вихід: щойно `check.sh` надрукував `PASS` — виконай `touch docs/features/board-protocol/execution/T13a/DONE` і завершуй.
Якщо та сама причина FAIL повторюється третю ітерацію поспіль і ти не можеш її усунути, не міняючи тест, —
запиши причину в `docs/features/board-protocol/execution/T13a/BLOCKED` і завершуй.

## Не роби

- Не змінюй приймальні тести; червоний тест — виправляй код.
- Не змінюй `CLAUDE.md`, story-файли, `check.sh`, `PROMPT.md`.
- Не стейдж нічого, крім файлів із пункту 6 (`git add <path>`, ніколи `git add -A` / `git add .`);
  файли в `docs/features/board-protocol/execution/` не комітиш.
- Не роби `git push`, не відкривай PR, не переписуй історію (`--amend` теж ні — помилка → новий коміт).
- Не додавай нових залежностей і нічого поза `governance/ports`, `governance/adapters/filesfake`,
  `governance/domain/errors.go` і трекером.
