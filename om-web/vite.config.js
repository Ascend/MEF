/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2022-2025. All rights reserved.
 */

import { fileURLToPath, URL } from 'node:url';

import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { defaultExclude } from 'vitest/config';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    target: 'esnext', // 浏览器可以处理最新的 ES 特性
    rollupOptions: {
      output: {
        // 将element-plus分到不同chunk
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (/element-plus\//.test(id)) {
              return 'element-plus';
            } else {
              return 'vendor';
            }
          }
          return undefined;
        },
        // 优化打包文件名
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId ? chunkInfo.facadeModuleId.split('/') : [];
          const fileName = facadeModuleId[facadeModuleId.length - 2] || '[name]';
          return `assets/${fileName}.[hash].js`;
        },
      },
    },
    chunkSizeWarningLimit: 1600,
  },
  test: {
    environment: 'happy-dom',
    reporter: ['verbose', 'junit'],
    outputFile: {
      junit: './test/reports/js_test.xml',
    },
    coverage: {
      reportsDirectory: './test/coverage',
      all: true,
      exclude: [
        ...defaultExclude,
        '**/public/**',
        '.*.cjs',
        '**/test/**',
        '**/router/index.js',
        'src/main.js',
      ],
    },
  },
});
