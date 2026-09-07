import React, { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { useDispatch } from 'react-redux';
import { setTheme } from '@/store/slices/appSlice';
import i18n, { persistLang, readLang, type AppLang } from '@/i18n';
import {
  applyDocumentTheme,
  persistThemePreference,
  readThemePreference,
  resolveTheme,
  type ResolvedTheme,
  type ThemePreference,
} from '@/theme/preference';
import type { AppDispatch } from '@/store';

interface ThemeLangValue {
  lang: AppLang;
  setLang: (lang: AppLang) => void;
  preference: ThemePreference;
  setPreference: (pref: ThemePreference) => void;
  resolved: ResolvedTheme;
}

const ThemeLangContext = createContext<ThemeLangValue | null>(null);

function useSystemDark(): boolean {
  const [dark, setDark] = useState(() => window.matchMedia('(prefers-color-scheme: dark)').matches);
  useEffect(() => {
    const mq = window.matchMedia('(prefers-color-scheme: dark)');
    const onChange = () => setDark(mq.matches);
    mq.addEventListener('change', onChange);
    return () => mq.removeEventListener('change', onChange);
  }, []);
  return dark;
}

export const ThemeLangProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const dispatch = useDispatch<AppDispatch>();
  const [lang, setLangState] = useState<AppLang>(readLang);
  const [preference, setPrefState] = useState<ThemePreference>(readThemePreference);
  const systemDark = useSystemDark();
  const resolved = resolveTheme(preference, systemDark);

  useEffect(() => {
    applyDocumentTheme(resolved);
    dispatch(setTheme(resolved));
  }, [dispatch, resolved]);

  const value = useMemo<ThemeLangValue>(() => ({
    lang,
    setLang: (next) => {
      persistLang(next);
      void i18n.changeLanguage(next);
      setLangState(next);
    },
    preference,
    setPreference: (next) => {
      persistThemePreference(next);
      setPrefState(next);
    },
    resolved,
  }), [lang, preference, resolved]);

  return <ThemeLangContext.Provider value={value}>{children}</ThemeLangContext.Provider>;
};

export function useThemeLang(): ThemeLangValue {
  const ctx = useContext(ThemeLangContext);
  if (!ctx) {
    throw new Error('useThemeLang must be used within ThemeLangProvider');
  }
  return ctx;
}
