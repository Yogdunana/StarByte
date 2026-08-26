import React from 'react';
import { useRoutes } from 'react-router-dom';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';

import routes from './router/routes';
import theme from './styles/theme';
import { ErrorBoundary } from './components';

const App: React.FC = () => {
  const element = useRoutes(routes);
  return (
    <ConfigProvider theme={theme} locale={zhCN}>
      <ErrorBoundary>
        {element}
      </ErrorBoundary>
    </ConfigProvider>
  );
};

export default App;
