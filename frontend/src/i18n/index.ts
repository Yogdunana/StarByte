import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import 'dayjs/locale/en';
import 'dayjs/locale/ru';
import zhCN from '@/locales/zh-CN.json';
import enUS from '@/locales/en-US.json';
import ruRU from '@/locales/ru-RU.json';
import zhText from '@/locales/zh-CN.text.json';
import enText from '@/locales/en-US.text.json';
import ruText from '@/locales/ru-RU.text.json';

export const LANG_KEY = 'starbyte_lang';
export type AppLang = 'zh-CN' | 'en-US' | 'ru-RU';

export function readLang(): AppLang {
  const raw = localStorage.getItem(LANG_KEY);
  return raw === 'en-US' || raw === 'ru-RU' ? raw : 'zh-CN';
}

export function persistLang(lang: AppLang): void {
  localStorage.setItem(LANG_KEY, lang);
  dayjs.locale(lang === 'ru-RU' ? 'ru' : lang === 'en-US' ? 'en' : 'zh-cn');
  if (typeof document !== 'undefined') {
    document.documentElement.lang = lang;
    document.title = { 'zh-CN': zhCN, 'en-US': enUS, 'ru-RU': ruRU }[lang].common.appTitle;
  }
}

void i18n.use(initReactI18next).init({
  resources: {
    'zh-CN': { translation: zhCN, text: zhText },
    'en-US': { translation: enUS, text: enText },
    'ru-RU': { translation: ruRU, text: ruText },
  },
  lng: typeof window === 'undefined' ? 'zh-CN' : readLang(),
  fallbackLng: 'zh-CN',
  interpolation: { escapeValue: false },
});

persistLang(readLang());

export default i18n;
