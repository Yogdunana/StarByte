import { describe, expect, it } from 'vitest';
import lightTheme, { darkComponents, darkTokens } from './theme';

describe('antd theme tokens', () => {
  it('uses a restrained ink accent instead of default Ant Design blue', () => {
    expect(lightTheme.token?.colorPrimary).toBe('#2f3530');
    expect(lightTheme.token?.colorBgLayout).toBe('#f3f2ee');
    expect(darkTokens?.colorPrimary).toBe('#d4d0c8');
    expect(darkTokens?.colorBgLayout).toBe('#121211');
    expect(lightTheme.components?.Button).toMatchObject({ primaryShadow: 'none' });
    expect(darkComponents?.Layout).toMatchObject({ siderBg: '#1c1c1a' });
  });
});
