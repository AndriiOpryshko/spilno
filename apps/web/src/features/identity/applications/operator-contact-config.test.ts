// @vitest-environment node
import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, readFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { describe, expect, it } from 'vitest';

const webRoot = path.resolve(import.meta.dirname, '../../../..');
const variable = 'VITE_OPERATOR_CONTACT_EMAIL';

describe('оператор contact address configuration (AC-05)', () => {
  it('documents the variable in apps/web/.env.example with a placeholder address', () => {
    const file = path.join(webRoot, '.env.example');
    expect(existsSync(file)).toBe(true);
    expect(readFileSync(file, 'utf8')).toMatch(new RegExp(`^${variable}=\\S+@\\S+`, 'm'));
  });

  it('fails a production build without the address and names the variable', () => {
    const env = { ...process.env };
    delete env[variable];
    const outDir = mkdtempSync(path.join(tmpdir(), 'spilno-build-'));
    const result = spawnSync('pnpm', ['exec', 'vite', 'build', '--mode', 'production', '--outDir', outDir], {
      cwd: webRoot,
      env,
      encoding: 'utf8',
    });
    expect(result.status).not.toBe(0);
    expect(`${result.stdout}${result.stderr}`).toContain(variable);
  }, 120_000);
});
