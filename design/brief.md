# Design brief: Cases (basic level)

Screen: Cases list

Current URL: http://localhost:5180/cases

Goal: the board chair sees at a glance which cases need attention and opens the protocol draft in one click.

Keep:

- search by case title filters the list as you type; `/` focuses the search field;
- status filter (All / Draft / In review / Ready to print / Printed) with counts;
- each case opens `/cases/$caseId/protocol`;
- shadcn/ui primitives in `apps/web/src/components/ui` (Button, Input, Skeleton, Kbd);
- data comes from `listCases()` through TanStack Query: `title`, `createdAt`, `documents[].kind`, `amountUah` (nullable), `status`.

States:

- default: list of cases;
- loading: skeleton rows;
- success: N/A (read-only screen);
- error: calm alert with "Спробувати ще раз";
- empty: "Немає справ за цим фільтром" with "Скинути фільтри";
- disabled: N/A.

Layouts:

- desktop: 1440px, dark sidebar + table with status glyphs;
- mobile: 390px, top bar + stacked case cards, filters scroll horizontally.

Design references:

- Figma file: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84
- current screenshot (prototype before redesign): design/screenshots/cases-before-desktop.png, design/screenshots/cases-before-mobile.png
- desktop frame: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=22-2
- mobile frame: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=22-118
- states frame: https://www.figma.com/design/D06IijJnLn8WEgU85fnD84?node-id=22-192
- Pencil source: N/A

Done:

- the approved design is implemented in `apps/web/src/features/governance/cases-screen.tsx`;
- the Keep behaviour works;
- `pnpm --filter @spilno/web lint`, `typecheck`, `test` pass;
- desktop and mobile checked in the browser (design/screenshots/cases-after-*.png);
- git diff touches only the governance feature, theme tokens and shared primitives it needs.
