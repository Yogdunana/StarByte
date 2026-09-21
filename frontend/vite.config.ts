import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');

  return {
    plugins: [react()],
    resolve: {
      alias: {
        // vite 8 起 `configLoader: 'native'` 不再提供 __dirname（会告警），
        // 官方建议改用 import.meta.dirname（Node >= 20.11，CI 的 Node 22 满足）。
        '@': path.resolve(import.meta.dirname, './src'),
      },
    },
    server: {
      port: 5173,
      host: true,
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://localhost:8080',
          changeOrigin: true,
        },
        '/ws': {
          target: env.VITE_WS_URL || 'ws://localhost:8080',
          ws: true,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: 'dist',
      sourcemap: mode !== 'production',
      rollupOptions: {
        output: {
          // vite 8 把打包器从 rollup 换成了 rolldown。rolldown 不支持 rollup 的
          // 「对象形式」manualChunks（会直接报 Invalid type: Expected Function but
          // received Object），而函数形式在 rolldown 里已标记 deprecated。
          // 这里改用官方推荐的 codeSplitting.groups：`test` 匹配模块，
          // includeDependenciesRecursively 默认 true，等价于 rollup 对象形式
          // 「把独占依赖一起收进该 chunk」的行为，因此分包结果与升级前一致。
          codeSplitting: {
            groups: [
              { name: 'react', test: /node_modules[\\/](react|react-dom|react-router-dom)[\\/]/ },
              { name: 'redux', test: /node_modules[\\/](@reduxjs[\\/]toolkit|react-redux)[\\/]/ },
              { name: 'antd', test: /node_modules[\\/](antd|@ant-design[\\/]icons)[\\/]/ },
              { name: 'echarts', test: /node_modules[\\/](echarts|echarts-for-react)[\\/]/ },
              // reactflow 只是个 re-export 壳，真正的实现在 @reactflow/* 下，
              // 只匹配 node_modules/reactflow/ 会漏掉全部实体模块。
              { name: 'reactflow', test: /node_modules[\\/](reactflow|@reactflow[\\/][^\\/]+)[\\/]/ },
            ],
          },
        },
      },
    },
    css: {
      modules: {
        localsConvention: 'camelCase',
      },
    },
  };
});
