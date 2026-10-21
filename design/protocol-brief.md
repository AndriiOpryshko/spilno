# Design brief: Protocol draft (advanced level)

Screen: Protocol draft editor

Current URL: http://localhost:5180/cases/c-101/protocol

Goal: the chair can tell in seconds which values came from the system, the AI or a human, where each number came from, and what still blocks printing; mistakes are caught while typing with a calm hint, not a red error.

Keep:

- system fields (number, date, board) are read-only and rendered from data;
- editable slots save through `updateSlot()`; saving makes the slot `origin: human`;
- non-editable AI slots (amounts) cannot be edited; the amount is also shown in words;
- tier-2 warnings are advisory, tier-3 checks block printing (`getReadiness()`);
- "Надрукувати протокол" stays disabled until there are no blocking checks and no unresolved slots;
- the same shadcn/ui primitives as the Cases screen.

Live validation (`validateSlot()`, runs on every keystroke, never blocks saving):

- value differs from what the documents say → amber hint with a one-click "Підставити";
- empty slot with known options → amber hint with example chips;
- empty slot without options → "Поле потрібне для протоколу".

Data per slot: `label`, `origin` (system / ai / human), `value`, `status` (verified / needs_review / not_found), `editable`, `sourceRef` (document title + fragment, nullable), optional `amount`, `expected`, `examples`.
Data per warning: `tier` (2 / 3), `title`, `message`, `slotId` (nullable), `ruleRef` (title + fragment).

States:

- default: draft with mixed slot statuses and a tier-3 limit;
- loading: skeleton;
- editing: "Редагується" tag, Save / Cancel, `⌘ Enter` saves and `Esc` cancels on desktop;
- slot saving: Save shows "Збереження…" and is disabled;
- slot not found: dashed field, example chips;
- tier-3 limit attached to its slot;
- error: draft not available, link back to Cases (`/cases/c-102/protocol`);
- ready: all slots verified, Print enabled.

Layouts:

- desktop: 1440px, sidebar + slot rows (label column 220px, value column) + side panel with readiness and properties;
- mobile: 390px, single column, readiness card above the slots, properties below.

Design references:

- Figma file: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84
- approved direction and explored variants: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=11-2
- current screenshot (prototype before redesign): design/screenshots/protocol-before-desktop.png, design/screenshots/protocol-before-mobile.png
- desktop frame: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=23-137
- mobile frame: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=23-420
- states frame: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=23-650
- dark mode check: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=23-739
- Pencil source: N/A

Done:

- the approved design is implemented in `apps/web/src/features/governance/protocol-screen.tsx` and `slot-row.tsx`;
- the Keep behaviour works, `governance.test.ts` stays green;
- `pnpm --filter @spilno/web lint`, `typecheck`, `test` pass;
- desktop and mobile checked in the browser (design/screenshots/protocol-after-*.png);
- diff layers: new values land in theme variables, the screen is composed from existing components, data enters through the props listed above.
