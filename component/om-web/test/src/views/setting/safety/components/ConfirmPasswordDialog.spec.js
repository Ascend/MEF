import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import ConfirmPasswordDialog from '@/views/setting/safety/components/ConfirmPasswordDialog.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const wrapper = shallowMount(ConfirmPasswordDialog, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety components ConfirmPasswordDialog onConfirm', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formRefMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  await wrapper.vm.onConfirm(formRefMock);
  expect(validateMock).toHaveBeenCalled();
});