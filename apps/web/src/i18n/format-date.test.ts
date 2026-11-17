import { readdirSync, readFileSync, statSync } from 'node:fs';
import path from 'node:path';
import { afterAll, beforeAll, describe, expect, it } from 'vitest';

import { formatDate } from './uk';

// A browser far from Kyiv: the formatter must not depend on the zone the user's machine is in.
const originalTz = process.env.TZ;
beforeAll(() => {
  process.env.TZ = 'America/Los_Angeles';
});
afterAll(() => {
  if (originalTz === undefined) delete process.env.TZ;
  else process.env.TZ = originalTz;
});

describe('formatDate (AC-05, AC-29): Europe/Kyiv calendar dates', () => {
  it('renders the instant 2026-11-10T22:30Z as 11.11.2026, the Kyiv day (UTC+2)', () => {
    expect(formatDate('2026-11-10T22:30:00Z')).toBe('11.11.2026');
  });

  it('keeps 2026-11-10T21:30Z on 10.11.2026, just before Kyiv midnight', () => {
    expect(formatDate('2026-11-10T21:30:00Z')).toBe('10.11.2026');
  });

  it('uses the summer offset (UTC+3): 2026-07-10T21:30Z is already 11.07.2026', () => {
    expect(formatDate('2026-07-10T21:30:00Z')).toBe('11.07.2026');
    expect(formatDate('2026-07-10T20:59:00Z')).toBe('10.07.2026');
  });

  it('formats a calendar date (term, extract date) without a zone shift', () => {
    expect(formatDate('2026-11-08')).toBe('08.11.2026');
    expect(formatDate('2026-01-01')).toBe('01.01.2026');
  });
});

function sourceFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const full = path.join(dir, name);
    if (full.endsWith(path.join('src', 'api', 'gen'))) return [];
    if (statSync(full).isDirectory()) return sourceFiles(full);
    return /\.(ts|tsx)$/.test(name) && !/\.(test|stories)\.tsx?$/.test(name) && !name.endsWith('.gen.ts') ? [full] : [];
  });
}

describe('one date formatter in the web app', () => {
  const root = path.resolve(import.meta.dirname, '..');
  const files = sourceFiles(root);
  const rel = (f: string) => path.relative(root, f);

  it('has no format-date module and no formatDay helper', () => {
    expect(files.map(rel).filter((f) => f.endsWith('format-date.ts'))).toEqual([]);
    expect(files.filter((f) => /\bformatDay\b/.test(readFileSync(f, 'utf8'))).map(rel)).toEqual([]);
  });

  it('has no toLocaleDateString and no ISO slicing outside i18n/uk.ts', () => {
    const offenders = files
      .filter((f) => rel(f) !== 'i18n/uk.ts')
      .filter((f) => /toLocaleDateString|\.slice\(\s*0\s*,\s*10\s*\)/.test(readFileSync(f, 'utf8')))
      .map(rel);
    expect(offenders).toEqual([]);
    expect(readFileSync(path.join(root, 'i18n/uk.ts'), 'utf8')).not.toMatch(/toLocaleDateString/);
    expect(readFileSync(path.join(root, 'i18n/uk.ts'), 'utf8')).toMatch(/timeZone:\s*'Europe\/Kyiv'/);
  });
});
