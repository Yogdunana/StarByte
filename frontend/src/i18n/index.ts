import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import 'dayjs/locale/en';
import zhCN from '@/locales/zh-CN.json';
import enUS from '@/locales/en-US.json';

export const LANG_KEY = 'starbyte_lang';
export type AppLang = 'zh-CN' | 'en-US';

export function readLang(): AppLang {
  const raw = localStorage.getItem(LANG_KEY);
  return raw === 'en-US' ? 'en-US' : 'zh-CN';
}

export function persistLang(lang: AppLang): void {
  localStorage.setItem(LANG_KEY, lang);
  dayjs.locale(lang === 'en-US' ? 'en' : 'zh-cn');
}

void i18n.use(initReactI18next).init({
  resources: {
    'zh-CN': { translation: zhCN },
    'en-US': { translation: enUS },
  },
  lng: typeof window === 'undefined' ? 'zh-CN' : readLang(),
  fallbackLng: 'zh-CN',
  interpolation: { escapeValue: false },
});

persistLang(i18n.language === 'en-US' ? 'en-US' : 'zh-CN');

export default i18n;
