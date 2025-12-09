import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as safety from '@/api/safety';
import * as commonMethods from '@/utils/commonMethods';
import EditLoginRuleDialog from '@/views/setting/safety/components/EditLoginRuleDialog.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const wrapper = shallowMount(EditLoginRuleDialog, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

const modifyLoginRulesMock = vi.spyOn(safety, 'modifyLoginRules');
const deepCopyDictListMock = vi.spyOn(commonMethods, 'deepCopyDictList');
modifyLoginRulesMock.mockResolvedValue({ data: 'mock-data' });
deepCopyDictListMock.mockResolvedValue({ data: 'mock-data' });

it('test setting safety components EditLoginRuleDialog onDialogConfirm', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formRefMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  await wrapper.vm.onDialogConfirm(formRefMock);
  expect(deepCopyDictListMock).toHaveBeenCalled();
});