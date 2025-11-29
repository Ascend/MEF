import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as safety from '@/api/safety';
import * as commonMethods from '@/utils/commonMethods';

import WeakPassword from '@/views/setting/safety/WeakPassword.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const exportPunyDictMock = vi.spyOn(safety, 'exportPunyDict');
const deletePunyDictMock = vi.spyOn(safety, 'deletePunyDict');
const importPunyDictMock = vi.spyOn(safety, 'importPunyDict');
const isValidUploadFilenameMock = vi.spyOn(commonMethods, 'isValidUploadFilename');
exportPunyDictMock.mockResolvedValue({ data: 'mock-data' });
deletePunyDictMock.mockResolvedValue({ data: 'mock-data' });
importPunyDictMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(WeakPassword, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety WeakPassword beforeWeakDictUpload', async () => {
  const rawFileMock = {
    name: 'nameMock',
    size: 32,
  };
  await wrapper.vm.beforeWeakDictUpload(rawFileMock);
  expect(isValidUploadFilenameMock).toHaveBeenCalledWith(rawFileMock.name);
});

it('test setting safety WeakPassword exportWeakDict', async () => {
  await wrapper.vm.exportWeakDict();
  expect(exportPunyDictMock).toHaveBeenCalled();
});

it('test setting safety WeakPassword onUploadSuccess', async () => {
  wrapper.vm.form = {
    fileName: 'xxxxx',
    password: '1233456789',
  }
  wrapper.vm.resetUploadForm = vi.fn();
  await wrapper.vm.onUploadSuccess();
  expect(importPunyDictMock).toHaveBeenCalled();
});

it('test setting safety WeakPassword submit', async () => {
  wrapper.vm.ops = {
    import: 'import',
    delete: 'delete',
  }
  wrapper.vm.submitImport = vi.fn();
  wrapper.vm.submitDelete = vi.fn();
  const formRefMock = {
    validate: vi.fn((callback) => {
      callback(true);
    }),
    resetFields: vi.fn(),
  }
  await wrapper.vm.submit(formRefMock);
});