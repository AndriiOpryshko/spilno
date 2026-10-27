# Design system

What the web app (`apps/web`) actually ships today. Every value below is read from code; nothing
here is a proposal except the [Gaps](#gaps) section. When code and this file disagree, code wins
and this file is stale.

- Theme: `apps/web/src/index.css` (tokens applied in commit `2b99ace` "feat(web): apply approved
  design tokens and primitives").
- Primitives: `apps/web/src/components/ui/` (shadcn, `cva` + `cn`, Radix `Slot` for `asChild`).
- Domain components: `apps/web/src/features/governance/`.
- Workshop: `mise run storybook` (:6006), config in `apps/web/.storybook/`.

## Sources

- Figma file: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84 ("Color" and "Radius" variable
  collections, modes Light / Dark; page "02 Components").
- Cases list frames: desktop `22-2`, mobile `22-118`, states `22-192` ([brief](brief.md)).
- Protocol draft frames: approved direction `11-2`, desktop `23-137`, mobile `23-420`, states
  `23-650`, dark mode check `23-739` ([protocol brief](protocol-brief.md)).
- Pencil source: N/A.

## Color tokens

Tailwind v4: each CSS variable is exposed as a utility through `@theme inline`
(`--color-<name>: var(--<name>)`), so `--warning` is used as `text-warning`, `bg-warning-subtle`,
etc. Dark mode is the `.dark` class on an ancestor (`@custom-variant dark (&:is(.dark *))`).

| Variable | Light (`:root`) | Dark (`.dark`) | Used for |
|---|---|---|---|
| `--background` | `#f7f7f9` | `#0f0f13` | page background (`body`) |
| `--foreground` | `#16161d` | `#ededf2` | body text |
| `--card` | `#ffffff` | `#17171d` | Card, outline Button, Kbd surface |
| `--card-foreground` | `#16161d` | `#ededf2` | text on card |
| `--popover` | `#ffffff` | `#17171d` | popovers (no consumer yet) |
| `--popover-foreground` | `#16161d` | `#ededf2` | |
| `--primary` | `#0f766e` | `#14b8a6` | primary Button, default Badge, links |
| `--primary-foreground` | `#ffffff` | `#04201d` | text on primary |
| `--secondary` | `#e6f7f5` | `#12302d` | tonal Button ("Підставити") |
| `--secondary-foreground` | `#0f766e` | `#5eead4` | |
| `--muted` | `#f1f1f4` | `#1f1f27` | disabled Button background |
| `--muted-foreground` | `#6e6e7c` | `#9a9aa8` | secondary text, placeholders, disabled text |
| `--accent` | `#f1f1f4` | `#1f1f27` | hover on ghost/outline, Skeleton |
| `--accent-foreground` | `#16161d` | `#ededf2` | |
| `--destructive` | `#b4374a` | `#f27185` | errors, tier-3 (blocking) checks, `StatusGlyph blocked` |
| `--destructive-subtle` | `#fbecee` | `#33161c` | background of blocking notes |
| `--success` | `#0f766e` | `#2dd4bf` | `StatusGlyph done` |
| `--warning` | `#935600` | `#f5b54a` | tier-2 hints, `StatusGlyph attention/empty` |
| `--warning-subtle` | `#fff4e0` | `#2e2210` | example chips in `InlineHint`, tier-2 `LimitNote` |
| `--origin-system` | `#5f5f6c` | `#9a9aa8` | `OriginTag system` text |
| `--origin-system-subtle` | `#f1f1f4` | `#24242c` | `OriginTag system` background |
| `--origin-ai` | `#5847e6` | `#a99bff` | `OriginTag ai` text |
| `--origin-ai-subtle` | `#efedff` | `#221f3d` | `OriginTag ai` background |
| `--origin-human` | `#1866ab` | `#6cb6f0` | `OriginTag human/editing` text |
| `--origin-human-subtle` | `#e6f1fb` | `#14263a` | `OriginTag human/editing` background |
| `--border` | `#e6e6eb` | `#2a2a33` | default border (`* { border-border }`) |
| `--input` | `#e6e6eb` | `#2a2a33` | Input / Textarea border |
| `--ring` | `#14b8a6` | `#2dd4bf` | focus ring (`ring-ring/50`, 3px) and outline |
| `--sidebar` | `#121217` | `#0a0a0d` | dark app sidebar (dark in both modes) |
| `--sidebar-foreground` | `#c9c9d3` | `#c9c9d3` | |
| `--sidebar-accent` | `#1e1e26` | `#1a1a21` | active sidebar item |
| `--brand-from` | `#efa58c` | (inherits light) | `BrandMark` gradient |
| `--brand-via` | `#8fd3c7` | (inherits light) | |
| `--brand-to` | `#14b8a6` | (inherits light) | |

Semantic meaning is fixed: **color encodes meaning, shape encodes progress** (`StatusGlyph`:
ring = state, dashed = empty/neutral, fill = done, dot = attention/blocked). Provenance of a value
(system / AI / human) always uses the `origin-*` pair, never status colors.

Dark mode: tokens exist for every color and Storybook switches them with a Theme toolbar
(`.storybook/preview.tsx` toggles `.dark` on `<html>`). The app itself has no theme switch yet.

## Radius

Overrides Tailwind's scale in `@theme inline`, mirrors the Figma "Radius" collection.

| Token | Value | Where |
|---|---|---|
| `--radius-sm` / `rounded-sm` | 6px | Kbd, OriginTag |
| `--radius-md` / `rounded-md` | 8px | Button (all sizes), Input, Textarea, Skeleton |
| `--radius-lg` / `rounded-lg` | 12px | panels: slot list, readiness card, skeleton blocks |
| `--radius-xl` / `rounded-xl` | 16px | Card |
| `rounded-full` | - | Badge |

## Typography

- Family: **Onest Variable** (`@fontsource-variable/onest` 5.3.1), `--font-sans: "Onest Variable",
  ui-sans-serif, system-ui, sans-serif`; `body` uses `font-sans antialiased`. Kbd also uses the
  sans family. No mono or display family.
- There are no type-scale tokens; sizes are Tailwind utilities or arbitrary values. In use:

| Size | Where |
|---|---|
| 11px (`text-[11px]`) | Kbd, small meta |
| 12px (`text-xs`) | Badge, OriginTag, captions |
| 13px (`text-[13px]`) | dense secondary text in governance screens (most frequent arbitrary size) |
| 14px (`text-sm`) | Button, Input/Textarea from `md`, body of panels |
| 15px (`text-[15px]`, `md:text-[15px]`) | slot values |
| 16px (`text-base`) | Input/Textarea below `md` (avoids iOS zoom), amount fraction |
| 22px → 28px (`text-[22px] leading-7 lg:text-[28px] lg:leading-9`) | protocol page `h1` |
| 24px (`text-2xl`) | home page `h1` |
| 26px, 28px → 32px | amount figures in `SlotRow` (`font-bold tracking-tight`) |

- Weights: 500 `font-medium` (Button, Badge, labels), 600 `font-semibold` (headings, CardTitle,
  OriginTag), 700 `font-bold` (amounts only).

## Spacing and layout

Tailwind's 4px scale; no custom spacing tokens. Conventions observed in code:

- Page gutter: `px-4` mobile, `lg:px-8` desktop; vertical `py-6` / `lg:py-8`.
- Inside panels and rows: `px-4` mobile, `md:px-6`; Card uses `py-6 px-6`, `gap-6`.
- Inline groups: `gap-2` (most common), `gap-1` / `gap-1.5` for icon + label, `gap-3`/`gap-4` between
  blocks; vertical stacks `space-y-2` … `space-y-6`.
- Breakpoints: Tailwind defaults; `md` (768px) switches slot rows to a `220px | 1fr` grid,
  `lg` (1024px) adds the sidebar and the 320px side panel (`lg:grid-cols-[1fr_320px]`).
- Reference frames: desktop 1440px, mobile 390px (briefs).
- Control heights: Button `h-9` (default), `h-8` (sm), `h-6` (xs), `h-10` (lg); Input `h-9`;
  Kbd and OriginTag `h-5`.

## Components

### Primitives (`src/components/ui`)

| Component | API / variants | Story | Used by |
|---|---|---|---|
| `Button` | `variant`: `default` (primary), `secondary` (tonal), `outline`, `ghost`, `link`, `destructive`; `size`: `default`, `xs`, `sm`, `lg`, `icon`, `icon-xs`, `icon-sm`, `icon-lg`; `asChild`. Disabled = neutral `bg-muted text-muted-foreground`, not opacity. | `UI/Button`: Primary, Tonal, Outline, Disabled | cases, protocol, slot row, readiness card |
| `Badge` | `variant`: `default`, `secondary`, `destructive`, `outline`, `ghost`, `link`; `asChild` | none | `features/system/health-card` |
| `Card` | `Card`, `CardHeader`, `CardTitle`, `CardDescription`, `CardAction`, `CardContent`, `CardFooter` | none | `features/system/health-card` |
| `Input` | native input props; focus ring, `aria-invalid` styling | none | cases search |
| `Textarea` | native textarea props; `field-sizing-content`, `min-h-16` | none (covered via `SlotRow` stories) | slot row editing |
| `Skeleton` | `animate-pulse bg-accent rounded-md` | none | cases, protocol loading |
| `Kbd` | `<kbd>` on `bg-card` (see Accessibility) | none (covered via `SlotRow` stories) | `/` hint in cases, `⌘ Enter` / `Esc` in slot row |

### Domain components (`src/features/governance`)

| Component | Variants / props | Story |
|---|---|---|
| `StatusGlyph` | `state`: `done`, `attention`, `empty`, `blocked`, `neutral`; optional `label` → `role="img"` | `Governance/Badges` → StatusGlyphs |
| `OriginTag` | `origin`: `system`, `ai`, `human`, `editing`; labels from `@/i18n/uk` | `Governance/Badges` → OriginTags |
| `InlineHint` | amber advisory hint with one-click "Підставити" / example chips | `Governance/InlineHint` |
| `LimitNote` | tier 2 (advisory) / tier 3 (blocking) note with `ruleRef` | `Governance/LimitNote` |
| `SlotRow` | slot with origin, status, source ref, edit / save / cancel | `Governance/SlotRow` |
| `ReadinessCard` | blocking checks, unresolved slots, Print button | `Governance/ReadinessCard` |
| `Sidebar`, `TopBar`, `BrandMark` (`src/components/app-shell.tsx`) | app shell, breadcrumbs | none |

Storybook titles mirror Figma "02 Components" names (comment `// Figma: ...` at the top of each
story file).

## Accessibility (enforced)

- Every story renders in Chromium through `@storybook/addon-vitest` and runs axe:
  `parameters.a11y.test = 'error'` in `.storybook/preview.tsx`, so a violation fails
  `pnpm --filter @spilno/web test:stories` (part of `mise run test`). Stories run in light mode by
  default; switch the Theme global to check dark.
- WCAG AA contrast is a hard rule. Precedent: commit `d181e56` moved `Kbd` from `bg-muted` to
  `bg-card` because `muted-foreground` on `muted` measured 4.45:1; on `card` it is 5.02:1. The
  Figma Kbd component was changed the same way. Do not put `muted-foreground` text on `muted`.
- Focus: all interactive primitives show `focus-visible:ring-[3px] ring-ring/50`; the base layer sets
  `outline-ring/50` everywhere.
- Invalid fields use `aria-invalid` (border and ring switch to `destructive`).
- Disabled buttons keep readable contrast (neutral fill instead of opacity).
- Status is never color-only: `StatusGlyph` changes shape and takes a text `label`; the loading
  skeleton region is `aria-busy`.
- Keyboard: `/` focuses cases search; `⌘ Enter` saves and `Esc` cancels slot editing.

## Usage rules

- Compose screens from `src/components/ui` primitives and existing feature components; new values go
  into theme variables in `index.css` (light and dark) and the Figma "Color" collection, never as
  hex in components.
- Add a primitive with the shadcn generator, then restyle it to the tokens; keep `data-slot`.
- Every new component variant gets a story next to the component (`*.stories.tsx`) that mirrors the
  Figma variants and passes the a11y test.
- Data: generated TanStack Query hooks from `@/api/gen/endpoints/<bc>/<bc>` (orval, `mise run gen`).
  `features/governance/api.ts` is still an in-memory stand-in with 350 ms latency until the
  governance contract lands.
- Pages live in `src/routes/` (TanStack file routes; never edit `routeTree.gen.ts`), feature
  components in `src/features/<bc>/`, UI strings in `src/i18n/uk.ts`.
- Never edit generated code (`src/api/gen/`, `routeTree.gen.ts`).

## Gaps

What the board-protocol Stage 1 screens (PRD US-01…US-13 Stage 1, SAD §6 Flows 1-13) need and the
code does not have yet. Each item is a candidate, not a decision.

### Tokens

- **Protocol lifecycle states** (`draft`, `ready_for_vote`, `ready_for_signing`, `signed`,
  `withdrawn`; UI labels чернетка, готовий до голосування, готовий до підпису, підписаний, знятий з
  розгляду): no state → color/glyph mapping. Today `CaseStatus` (`draft/review/ready/printed`) is the
  prototype's vocabulary and does not match.
- **`--success-subtle`** (and an info/neutral-subtle pair): `warning-subtle` and `destructive-subtle`
  exist, success does not; a "signed" banner or badge needs it.
- **Locked / read-only surface** for a signed protocol (Flow 5-6, AC-17): no token or pattern for
  immutable content.
- **Type scale tokens**: sizes are arbitrary values (13px, 15px, 22px, 28px); a named scale would
  stop drift.
- **Shadow / elevation** for overlays (dialog, popover, toast): only `shadow-xs` / `shadow-sm` in use.
- **Dark mode in the app**: tokens exist, but no switch and no `prefers-color-scheme` wiring.
- **Print styles**: printing is `window.print()` of the screen; the Stage 1 print copy is rendered
  server-side (Flow 4), so the screen needs a "print copy ready" pattern, not print CSS.

### Components

- **Protocol state badge**: a `Badge` variant (or `StatusGlyph` + label) per lifecycle state; `Badge`
  has no status variants and no story.
- **Lifecycle stepper**: чернетка → готовий до голосування → готовий до підпису → підписаний, with
  знятий з розгляду as an exit (Flows 2-5, 9) and "return to чернетка" (Flow 3).
- **Dialog / AlertDialog**: withdraw with a mandatory reason (AC-35…AC-37), confirm freezing for vote
  or signing, return to чернетка voiding the print copy. `radix-ui` is installed, no wrapper exists.
- **File upload / dropzone** with progress and error: signed scan (Flow 5, AC-14, AC-34) and case
  documents attached to the conversation (Flow 1).
- **Form set**: `Label`, `Select` (meeting form, signers), date input (meeting date, member term),
  `RadioGroup` (poll vote per agenda item, Flow 2), `Checkbox`; numeric inputs for meeting facts
  (attendance, vote counts, Flow 5).
- **Table** with keyset "load more": roster (Flow 12), реєстр протоколів search results (Flow 13);
  the cases list is a custom layout, not a reusable table.
- **Toast** for async outcomes and Google service failures (Flow 8, `governance.*` problem+json).
- **Alert / problem banner** (today the `role="alert"` block is hand-written twice, in `cases-screen` and `protocol-screen`) that renders RFC 9457 errors with a next action (e.g.
  `governance.settings_missing` with a link to settings, Flow 1).
- **Conversation thread** for the помічник протоколу (Flow 1, US-12): message list, attachments, a
  pending-job state while the worker proposes a draft.
- **Tabs or sectioned detail view** for an opened decision: signed copy, meeting facts, poll history,
  cited documents (Flow 13, US-07).
- **Document link with version notice**: a cited document that has a newer version (US-04).
- **Poll results summary** per agenda item and round (раунд опитування), labelled as not the
  meeting's vote.
- **Stories missing** for existing primitives: `Badge`, `Card`, `Input`, `Textarea`, `Skeleton`, `Kbd`,
  and the app shell; their a11y is only tested indirectly.
