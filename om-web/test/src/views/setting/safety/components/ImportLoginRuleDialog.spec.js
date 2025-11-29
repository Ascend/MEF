import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as commonMethods from '@/utils/commonMethods';
import * as safety from '@/api/safety';

import ImportLoginRuleDialog from '@/views/setting/safety/components/ImportLoginRuleDialog.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const isValidUploadFilenameMock = vi.spyOn(commonMethods, 'isValidUploadFilename');
const handleUploadErrorMock = vi.spyOn(commonMethods, 'handleUploadError');
const importLoginRulesMock = vi.spyOn(safety, 'importLoginRules');
handleUploadErrorMock.mockResolvedValue({ data: 'mock-data' });
importLoginRulesMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(ImportLoginRuleDialog, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety ImportLoginRuleDialog beforeLoginRuleUpload', async () => {
  const rawFileMock = {
    name: 'nameMock',
    size: 32,
  };
  wrapper.vm.formRef = {
    resetFields: vi.fn(),
  };
  await wrapper.vm.beforeLoginRuleUpload(rawFileMock);
  expect(isValidUploadFilenameMock).toHaveBeenCalledWith(rawFileMock.name);
});

it('test setting safety ImportLoginRuleDialog onUploadSuccess', async () => {
  await wrapper.vm.onUploadSuccess();
  expect(importLoginRulesMock).toHaveBeenCalled();
});

it('test setting safety ImportLoginRuleDialog onUploadError', async () => {
  wrapper.vm.upload = {
    clearFiles: vi.fn(),
  }
  await wrapper.vm.onUploadError();
  expect(handleUploadErrorMock).toHaveBeenCalled();
});