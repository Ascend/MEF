import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as account from '@/api/account';
import WebTimeout from '@/views/setting/safety/WebTimeout.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const querySessionServiceMock = vi.spyOn(account, 'querySessionService');
const modifySessionServiceMock = vi.spyOn(account, 'modifySessionService');
querySessionServiceMock.mockResolvedValue({ data: 'mock-data' });
modifySessionServiceMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(WebTimeout, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety WebTimeout submitChange', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.webTimeoutForm.SessionTimeout = 20;
  wrapper.vm.webTimeoutForm.Password = '12345678';
  await wrapper.vm.submitChange(formElementMock);
  expect(validateMock).toHaveBeenCalled();
});