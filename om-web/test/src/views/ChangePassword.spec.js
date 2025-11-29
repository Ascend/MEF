import { expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import * as common from '@/api/common';

import { createRouter, createWebHistory } from 'vue-router';

import ChangePassword from '@/views/ChangePassword.vue';

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
const wrapper = mount(ChangePassword, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('should test input', async () => {
  const pwdMock = 'mock_pwd';
  const input = wrapper.find('input[placeholder="请再次输入新密码"]');
  await input.setValue(pwdMock);
  expect(input.element.value).toBe('mock_pwd');

  fetchConfigJsonMock.mockRestore();
});

it('should text isLowPasswordStrengthLevel', () => {
  const passwordStrengthLevelMock = 'low';
  wrapper.vm.passwordStrengthLevel = passwordStrengthLevelMock;
  const isLowPasswordStrengthLevelMock = wrapper.vm.isLowPasswordStrengthLevel();
  expect(isLowPasswordStrengthLevelMock).toBe(true);
});

it('should text isLowPasswordStrengthLevel', () => {
  const passwordStrengthLevelMock = 'medium';
  wrapper.vm.passwordStrengthLevel = passwordStrengthLevelMock;
  const isMediumPasswordStrengthLevelMock = wrapper.vm.isMediumPasswordStrengthLevel();
  expect(isMediumPasswordStrengthLevelMock).toBe(true);
});

it('should text isLowPasswordStrengthLevel', () => {
  const passwordStrengthLevelMock = 'high';
  wrapper.vm.passwordStrengthLevel = passwordStrengthLevelMock;
  const isHighPasswordStrengthLevelMock = wrapper.vm.isHighPasswordStrengthLevel();
  expect(isHighPasswordStrengthLevelMock).toBe(true);
});