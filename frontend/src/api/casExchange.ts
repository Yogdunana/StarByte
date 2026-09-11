import type { CASExchangeResponse } from '@/types/api';

import { exchangeCasCode } from './auth';

const exchangeByCode = new Map<string, Promise<CASExchangeResponse>>();

export function exchangeCasCodeOnce(code: string): Promise<CASExchangeResponse> {
  let pending = exchangeByCode.get(code);
  if (!pending) {
    pending = exchangeCasCode(code);
    exchangeByCode.set(code, pending);
  }
  return pending;
}

export function resetCasCodeExchange(): void {
  exchangeByCode.clear();
}
