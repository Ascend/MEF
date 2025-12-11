/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { defaultExclude } from 'vitest/config';

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url))
    }
  },
  build: {
    target: 'esnext', // 浏览器可以处理最新的 ES 特性
    rollupOptions: {
      output: {
        // 将element-plus分到不同chunk
        manualChunks(id) {
          if (id.includes("node_modules")) {
            if (/element-plus\//.test(id)) {
              return "element-plus";
            } else {
              return "vendor";
            }
          }
        },
        // 优化打包文件名
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        chunkFileNames: (chunkInfo) => {
          const facadeModuleId = chunkInfo.facadeModuleId ? chunkInfo.facadeModuleId.split("/") : [];
          const fileName = facadeModuleId[facadeModuleId.length - 2] || "[name]";
          return `assets/${fileName}.[hash].js`;
        }
      }
    }
  },
  test: {
    environment: 'happy-dom',
    reporter: ['verbose', 'junit'],
    outputFile: {
      'junit': './test/reports/js_test.xml',
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
