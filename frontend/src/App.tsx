import React from 'react';
import { useRoutes } from 'react-router-dom';
import routes from './router/routes';

const App: React.FC = () => {
  const element = useRoutes(routes);
  return <>{element}</>;
};

export default App;
