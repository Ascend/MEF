import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import * as update from '@/api/update';
import * as information from '@/api/information';

import { createRouter, createWebHistory } from 'vue-router';

import SystemUpdate from '@/views/manager/update/SystemUpdate.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const queryUpdateStatusMock = vi.spyOn(update, 'queryUpdateStatus');
const updateFirmwareMock = vi.spyOn(update, 'updateFirmware');
const resetFirmwareMock = vi.spyOn(update, 'resetFirmware');

const querySystemsSourceInfoMock = vi.spyOn(information, 'querySystemsSourceInfo');

queryUpdateStatusMock.mockResolvedValue({ data: 'mock-data' });
updateFirmwareMock.mockResolvedValue({ data: 'mock-data' });
resetFirmwareMock.mockResolvedValue({ data: 'mock-data' });

querySystemsSourceInfoMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(SystemUpdate, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test manager update SystemUpdate onUploadChange', async () => {
  const fileMock = {
    name: 'mockName',
    status: 'ready',
  };
  wrapper.vm.onUploadChange(fileMock);
  expect(wrapper.vm.updateForm.updateFile).toBe('mockName')
});