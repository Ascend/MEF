import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as account from '@/api/account';
import PasswordValidity from '@/views/setting/safety/PasswordValidity.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const queryAccountServiceMock = vi.spyOn(account, 'queryAccountService');
const modifyAccountServiceMock = vi.spyOn(account, 'modifyAccountService');
const getUserInfoMock = vi.spyOn(account, 'getUserInfo');
queryAccountServiceMock.mockResolvedValue({ data: 'mock-data' });
modifyAccountServiceMock.mockResolvedValue({ data: 'mock-data' });
getUserInfoMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(PasswordValidity, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety PasswordValidity submitChange', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.passwordValidityForm.PasswordExpirationDays = 20;
  wrapper.vm.passwordValidityForm.Password = '12345678';
  await wrapper.vm.submitChange(formElementMock);
  expect(validateMock).toHaveBeenCalled();
});