import { Select } from 'antd';
import { useTranslation } from 'react-i18next';
import { tx } from '@/i18n/text';
import { useThemeLang } from '@/theme/ThemeLangContext';

export default function LanguageSelect() {
  const { lang, setLang } = useThemeLang();
  const { t } = useTranslation();
  return (
    <Select
      aria-label={t('topbar.language')}
      value={lang}
      onChange={setLang}
      style={{ minWidth: 120 }}
      options={[
        { value: 'zh-CN', label: tx('简体中文') },
        { value: 'en-US', label: 'English' },
        { value: 'ru-RU', label: 'Русский' },
      ]}
    />
  );
}
