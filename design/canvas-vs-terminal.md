# Canvas vs terminal: how the work split

Route: Official Figma MCP for both screens; the advanced level is the second, harder screen
(Protocol draft: slot provenance, live validation, blocking checks, readiness).

## Steps that were more natural on the canvas

- **Finding the form.** Four visual directions (document sheet, calm panel, bold dark,
  hybrid) were built side by side on one page and rejected or kept by looking at them. Each
  rejection ("too bureaucratic", "gradient text reads badly") took one glance; in code each
  would have been a rebuild of real components.
- **Comparing variants of one decision.** Button treatments and accent palettes were shown as
  identical mini previews with contrast numbers next to them. The choice was made by
  comparing, not by reasoning about hex values.
- **States at once.** Every slot state (verified, attention, editing, empty, locked) and every
  screen state (loading, error, empty, ready) sits on one board. In the browser only one state
  is visible at a time and most need data or timing to reach.
- **Token decisions.** Light and dark modes were checked by switching a frame's mode, before
  any CSS existed.

## Steps that were more natural in the terminal

- **Implementation against existing code.** The screen had to fit the repo's router,
  query client, shadcn primitives and folder rules; only the codebase knows those.
- **Behaviour.** Live validation, Save / Cancel, `⌘ Enter`, readiness recalculation are logic;
  they are unit-tested (`governance.test.ts`) and exercised by a browser script, which a canvas
  cannot do.
- **Verification.** Lint, typecheck, tests and a scripted browser pass on 1440 and 390 give
  a repeatable yes/no; a screenshot only gives an impression.
- **Reading the diff.** Whether new values landed in variables, whether the screen reuses
  existing components, and whether data enters through typed props is answered by the diff,
  layer by layer.

## Why

The canvas is cheap to change and shows many alternatives at once, so it wins while the
question is "what should this look like". The terminal is where the answer meets real
constraints (existing code, data, behaviour) and can be checked mechanically, so it wins once
the question becomes "does this work and is it the right change".
