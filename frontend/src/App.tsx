import React from 'react';
import { useRoutes } from 'react-router-dom';
import { ConfigProvider, theme as antdTheme } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import enUS from 'antd/locale/en_US';
import ruRU from 'antd/locale/ru_RU';

import routes from './router/routes';
import lightTheme, { darkComponents, darkTokens } from './styles/theme';
import { ErrorBoundary } from './components';
import { ThemeLangProvider, useThemeLang } from './theme/ThemeLangContext';
import MotionRoot from './motion/MotionRoot';

const ThemedApp: React.FC = () => {
  const element = useRoutes(routes);
  const { resolved, lang } = useThemeLang();
  const locale = lang === 'ru-RU' ? ruRU : lang === 'en-US' ? enUS : zhCN;
  const algorithm = resolved === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm;
  const isDark = resolved === 'dark';

  return (
    <ConfigProvider
      locale={locale}
      theme={{
        algorithm,
        token: isDark ? darkTokens : lightTheme.token,
        components: isDark ? darkComponents : lightTheme.components,
      }}
    >
      <ErrorBoundary>{element}</ErrorBoundary>
    </ConfigProvider>
  );
};

const App: React.FC = () => (
  <ThemeLangProvider>
    <MotionRoot>
      <ThemedApp />
    </MotionRoot>
  </ThemeLangProvider>
);

export default App;
