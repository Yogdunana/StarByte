import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { describe, expect, it } from 'vitest';

const dir = dirname(fileURLToPath(import.meta.url));
const source = readFileSync(join(dir, 'MainLayout.tsx'), 'utf8');
const css = readFileSync(join(dir, 'MainLayout.module.css'), 'utf8');

describe('MainLayout motion chrome', () => {
  it('fades brand copy but unmounts search and footer immediately', () => {
    expect(source).toContain('key="brand-copy"');
    expect(source).toContain('exit="hidden"');
    expect(source).not.toContain('key="menu-search"');
    expect(source).not.toContain('key="sidebar-footer"');
    expect(source).toMatch(/Enter-only:[\s\S]*menuSearch/);
  });

  it('wraps Outlet so fragment pages are not flex-grown as siblings', () => {
    expect(source).toMatch(/contentPage[\s\S]*<Outlet \/>/);
    expect(css).not.toMatch(/\.contentStage\s*>\s*\*/);
    expect(css).toContain('.contentPage');
  });
});
