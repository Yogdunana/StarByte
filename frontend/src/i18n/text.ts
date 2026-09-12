import { useTranslation } from 'react-i18next';
import i18n from './index';

/** Translate UI copy, keeping user-authored values as interpolation data. */
export function tx(source: string, values?: Record<string, unknown>): string {
  return String(i18n.t(source, { ns: 'text', keySeparator: false, nsSeparator: false, ...values }));
}
export function useLocale(): string {
  const { i18n: instance } = useTranslation('text');
  return instance.language;
}
