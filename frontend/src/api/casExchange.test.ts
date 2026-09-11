import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('./auth', () => ({
  exchangeCasCode: vi.fn(),
}));

import { exchangeCasCode } from './auth';
import { exchangeCasCodeOnce, resetCasCodeExchange } from './casExchange';

describe('exchangeCasCodeOnce', () => {
  beforeEach(() => {
    resetCasCodeExchange();
    vi.mocked(exchangeCasCode).mockReset();
  });

  it('reuses the first redeem for the same code', async () => {
    const payload = {
      access_token: 'a',
      refresh_token: 'r',
      redirect: '/dashboard',
    };
    vi.mocked(exchangeCasCode).mockResolvedValue(payload as never);

    const first = exchangeCasCodeOnce('code-1');
    const second = exchangeCasCodeOnce('code-1');
    await expect(first).resolves.toEqual(payload);
    await expect(second).resolves.toEqual(payload);
    expect(exchangeCasCode).toHaveBeenCalledTimes(1);
  });
});
