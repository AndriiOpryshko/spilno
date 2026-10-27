# First screen: protocol draft

The first screen of the board-protocol module is the **protocol draft editor**. Every Stage 1
flow passes through it: a draft is created into it (SAD Flow 1), frozen for the poll or for
signing from it (Flows 2, 4), returned to it (Flow 3) and withdrawn from it (Flow 9). The cases
list (`/cases`) only leads here.

- Route: `/cases/$caseId/protocol` → `apps/web/src/routes/cases.$caseId.protocol.tsx`, which renders
  `ProtocolScreen` keyed by `caseId`.
- Screen: `apps/web/src/features/governance/protocol-screen.tsx`.
- Local URL: http://localhost:5180/cases/c-101/protocol (`c-101` is the only draft in
  `features/governance/mock-data.ts`; any other id, e.g. `c-102`, shows the error state).
- Brief: [../protocol-brief.md](../protocol-brief.md). Design system: [../design-system.md](../design-system.md).
- Data: in-memory stand-in `features/governance/api.ts` (`getProtocol`, `updateSlot`, 350 ms latency)
  until the governance OpenAPI contract is generated; the screen does not need the core API yet.

Other routes today: `/` (health card), `/cases` (cases list), `/cases/$caseId/protocol`.

## What it shows

- **Top bar** with breadcrumbs (Справи → case title → protocol number) and a "saved / saved just now" indicator.
- **Header**: protocol number and case title; a properties block (date via `formatDate`, etc.) sits in the side panel / below the slots on mobile.
- **Slot rows**, one per decision value: label column (220px from `md`), value, `OriginTag`
  (system / AI / human / editing), `StatusGlyph` (verified / needs review / not found), source
  reference (document title + fragment), amounts with the sum in words.
- **Editing** an editable slot: `Textarea`, Save / Cancel, `⌘ Enter` / `Esc` (`Kbd`), live
  validation through `InlineHint` (one-click "Підставити", example chips, "Поле потрібне для
  протоколу"); saving makes the slot `origin: human`.
- **Warnings**: tier-2 advisory and tier-3 blocking `LimitNote`s attached to their slot.
- **Readiness card** (side panel at `lg`, above the slots on mobile): blocking checks, unresolved
  slots and "Надрукувати протокол", disabled until nothing blocks.
- **States**: loading skeleton (`aria-busy`), error (`role="alert"` with a link back to Cases),
  ready (Print enabled).

## Design-system pieces used

| Piece | From |
|---|---|
| `Button` (primary, secondary, outline `asChild` link) | `components/ui/button.tsx` |
| `Textarea`, `Kbd`, `Skeleton` | `components/ui/` |
| `TopBar` | `components/app-shell.tsx` |
| `SlotRow`, `OriginTag`, `StatusGlyph`, `InlineHint`, `LimitNote`, `ReadinessCard` | `features/governance/` |
| Tokens: `origin-*`, `warning`/`warning-subtle`, `destructive`/`destructive-subtle`, `success`, `card`, `border`, `ring`; radius `md`/`lg` | `apps/web/src/index.css` |

What the Stage 1 version of this screen will additionally need (state badge, lifecycle stepper,
dialogs, upload, toast) is listed in [Gaps](../design-system.md#gaps).

## Screenshots

From the design iteration (commit `b7a5d21`), 1440px desktop and 390px mobile:

| | Desktop | Mobile |
|---|---|---|
| Before (prototype) | ![](../screenshots/protocol-before-desktop.png) | ![](../screenshots/protocol-before-mobile.png) |
| After (approved design) | ![](../screenshots/protocol-after-desktop.png) | ![](../screenshots/protocol-after-mobile.png) |
| After, with edits | ![](../screenshots/protocol-after-edits-desktop.png) | ![](../screenshots/protocol-after-edits-mobile.png) |

## verify-ui results

_To be filled after `/sdlc:verify-ui` runs against http://localhost:5180/cases/c-101/protocol with
the Definition of Done from `docs/design/design-spec.md`._

| Check | Verdict | Expected vs actual | Evidence |
|---|---|---|---|
| | | | |

Summary: _pending_. Screenshots: `docs/design/first-screen/screenshots/`.
