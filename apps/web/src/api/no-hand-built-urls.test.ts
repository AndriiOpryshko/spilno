import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';
import { describe, expect, it } from 'vitest';

const src = join(__dirname, '..');

function sources(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    if (path === join(src, 'api', 'gen') || name === 'routeTree.gen.ts') return [];
    if (statSync(path).isDirectory()) return sources(path);
    const isCode = /\.tsx?$/.test(name) && !/\.(test|stories)\.tsx?$/.test(name);
    return isCode ? [path] : [];
  });
}

describe('API links (AC-04d, AC-05)', () => {
  it('spells no API path by hand outside api/gen', () => {
    const offenders = sources(src)
      .filter((file) => /['"`]\/(api|identity)\//.test(readFileSync(file, 'utf8')))
      .map((file) => relative(src, file));
    expect(offenders).toEqual([]);
  });
});
