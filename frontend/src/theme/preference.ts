export const THEME_KEY = 'starbyte_theme';
export type ThemePreference = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';

export function readThemePreference(): ThemePreference {
  const raw = localStorage.getItem(THEME_KEY);
  if (raw === 'dark' || raw === 'system') return raw;
  return 'light';
}

export function persistThemePreference(pref: ThemePreference): void {
  localStorage.setItem(THEME_KEY, pref);
}

export function resolveTheme(pref: ThemePreference, systemDark = false): ResolvedTheme {
  if (pref === 'system') return systemDark ? 'dark' : 'light';
  return pref;
}

export function applyDocumentTheme(theme: ResolvedTheme): void {
  document.documentElement.setAttribute('data-theme', theme);
}
