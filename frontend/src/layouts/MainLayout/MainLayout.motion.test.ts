import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const source = readFileSync(join(dirname(fileURLToPath(import.meta.url)), 'MainLayout.tsx'), 'utf8');

describe('MainLayout motion chrome', () => {
  it('fades brand copy but unmounts search and footer immediately', () => {
    expect(source).toContain('key="brand-copy"');
    expect(source).toContain('exit="hidden"');
    expect(source).not.toContain('key="menu-search"');
    expect(source).not.toContain('key="sidebar-footer"');
    expect(source).toMatch(/Enter-only:[\s\S]*menuSearch/);
  });
});
