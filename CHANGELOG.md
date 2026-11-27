# Changelog

All notable changes to spilno are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
(Added, Changed, Deprecated, Removed, Fixed, Security; an empty category is left out),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
The version lives in `VERSION`; `mise run version:next` derives the next one from the
Conventional Commits since the last `v*` tag. Friendly, user-facing notes for each release
are in `docs/releases/`.

After 0.2.0, merging a release PR runs `.github/workflows/release.yml`: it builds the
release's changelog section from the merged history (git-cliff, `cliff.toml`), tags the merge
commit and publishes it as a [GitHub Release](https://github.com/AndriiOpryshko/spilno/releases).
`main` is protected, so the workflow does not commit here: the Releases page is the complete
log, this file holds the hand-curated entries up to 0.2.0.

## [Unreleased]

## [0.2.0] - 2026-11-27

### Added

- Several checkouts of the repo can run side by side: each git worktree gets its own API and
  web ports, database and dependencies (`mise run wt:new`), and `WEB_PORT` moves the web dev
  server off the default 5180

### Fixed

- A contract counterparty made only of spaces or control characters is now rejected with a
  field error, while a name with surrounding spaces stays valid

## [0.1.0] - 2026-11-17

First tagged baseline: everything merged up to the end of association onboarding.

### Added

- Board meeting protocols: a draft protocol with agenda items and decision slots, poll rounds
  with the own-vote rule, readiness and revision checks, a canonical snapshot with its
  fingerprint, and the signed protocol view
- Association onboarding: filing an application, operator review with competing applications,
  the association roster and registry, cases with document upload and versions on Google Drive,
  and a notice for a disabled association
- A Google Drive connection for an association, with the refresh token sealed at rest
- Contracts and meetings as the first vertical slices of the finance and governance areas
- Sign-in through the development shim, a public status page and a favicon
- A design system with Storybook stories that run as visual and accessibility tests

[Unreleased]: https://github.com/AndriiOpryshko/spilno/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/AndriiOpryshko/spilno/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AndriiOpryshko/spilno/releases/tag/v0.1.0
