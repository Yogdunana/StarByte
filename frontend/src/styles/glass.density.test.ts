import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

describe('dialog density', () => {
  const css = readFileSync(join(dirname(fileURLToPath(import.meta.url)), 'glass.css'), 'utf8');

  it('keeps modal footers on short laptop viewports', () => {
    expect(css).toContain('@media (max-height: 900px)');
    expect(css).toContain('max-height: calc(100dvh - 24px)');
    expect(css).toContain('overflow-y: auto');
  });
});
