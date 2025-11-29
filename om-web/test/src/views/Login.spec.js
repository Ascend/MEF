import { expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import * as common from '@/api/common';

import { createRouter, createWebHistory } from 'vue-router';

import Login from '@/views/Login.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const fetchConfigJsonMock = vi.spyOn(common, 'fetchJson');
const mockResponse = {
  'model': 'Atlas 500 A2',
  'systemName': {
    'zh': 'mock data',
    'en': 'mock data',
  },
  'websiteTitle': {
    'zh': 'mock data',
    'en': 'mock data',
  },
  'userGuide': {
    'zh': 'mock data',
    'en': 'mock data',
  },
  'copyRight': {
    'zh': 'mock data',
    'en': 'mock data',
  },
};

fetchConfigJsonMock.mockResolvedValue(mockResponse);
const wrapper = mount(Login, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test Login.vue', async () => {
  const pwdMock = 'mock_pwd';
  const input = wrapper.find('input')
  await input.setValue(pwdMock)
  expect(input.element.value).toBe('mock_pwd')

  fetchConfigJsonMock.mockRestore();
});

it('test username input box cannot be empty', async () => {
  const loginFormRefMock = {'data': 'mock data'};
  const userNameMock = '';
  wrapper.vm.loginForm.UserName = userNameMock;
  wrapper.vm.submitLogin(loginFormRefMock);
  expect(wrapper.vm.checkUsername.errorText).toBe('输入不能为空')
});