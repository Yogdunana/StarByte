import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

describe('login density', () => {
  const css = readFileSync(join(dirname(fileURLToPath(import.meta.url)), 'Login.module.css'), 'utf8');

  it('lets the login shell scroll at 100% instead of clipping the form', () => {
    expect(css).toContain('overflow-y: auto');
    expect(css).toContain('min-height: 100dvh');
    expect(css).not.toMatch(/\.container \{[^}]*overflow:\s*hidden/);
  });

  it('tightens the card on short laptop viewports', () => {
    expect(css).toContain('@media (max-height: 900px)');
    expect(css).toContain('cardHeaderTop');
    expect(css).toContain('min-height: 38px');
  });
});
