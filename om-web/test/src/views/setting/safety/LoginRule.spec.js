import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as safety from '@/api/safety';
import * as commonMethods from '@/utils/commonMethods';
import LoginRule from '@/views/setting/safety/LoginRule.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const queryLoginRulesMock = vi.spyOn(safety, 'queryLoginRules');
const deepCopyDictListMock = vi.spyOn(commonMethods, 'deepCopyDictList');
queryLoginRulesMock.mockResolvedValue({ data: 'mock-data' });
deepCopyDictListMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(LoginRule, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety LoginRule onConfirmPasswordDeleteSubmit', async () => {
  const passwordMock = '12345678';
  await wrapper.vm.onConfirmPasswordDeleteSubmit(passwordMock);
  expect(deepCopyDictListMock).toHaveBeenCalled();
  const isShowConfirmPasswordDialogMock = wrapper.vm.isShowConfirmPasswordDialog;
  expect(isShowConfirmPasswordDialogMock).toBe(false)
});

it('test setting safety LoginRule refreshTable', async () => {
  await wrapper.vm.refreshTable();
  expect(queryLoginRulesMock).toHaveBeenCalled();
});

it('test setting safety LoginRule onConfirmPasswordSwitchSubmit', async () => {
  const passwordMock = '12345678';
  await wrapper.vm.onConfirmPasswordDeleteSubmit(passwordMock);
  expect(deepCopyDictListMock).toHaveBeenCalled();
  const isShowConfirmPasswordDialogMock = wrapper.vm.isShowConfirmPasswordDialog;
  expect(isShowConfirmPasswordDialogMock).toBe(false)
});
