import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/test/setup.ts'],
    css: true,
    // antd + RTL 的组件测试在 jsdom 下本来就慢（单个表单用例约 3s），
    // 并行跑整套时会被 CPU 争抢挤过 5s 默认上限而假失败。
    // 给足余量，避免把「机器忙」误报成「代码坏」；真正卡死仍会在 15s 内失败。
    testTimeout: 15000,
    hookTimeout: 15000,
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      include: ['src/components/FormEngine/**', 'src/components/StatusTag/**', 'src/components/EmptyState/**', 'src/components/DataTable/**'],
    },
  },
});
