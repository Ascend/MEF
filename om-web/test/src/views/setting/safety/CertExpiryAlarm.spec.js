import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as safety from '@/api/safety';
import CertExpiryAlarm from '@/views/setting/safety/CertExpiryAlarm.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const queryHttpsCertAlarmTimeMock = vi.spyOn(safety, 'queryHttpsCertAlarmTime');
const modifyHttpsCertAlarmTimeMock = vi.spyOn(safety, 'modifyHttpsCertAlarmTime');
queryHttpsCertAlarmTimeMock.mockResolvedValue({ data: 'mock-data' });
modifyHttpsCertAlarmTimeMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(CertExpiryAlarm, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety CertExpiryAlarm submitChange', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.certAlarmForm.CertAlarmTime = 20;
  wrapper.vm.certAlarmForm.Password = '12345678';
  await wrapper.vm.submitChange(formElementMock);
  expect(validateMock).toHaveBeenCalled();
  expect(modifyHttpsCertAlarmTimeMock).toHaveBeenCalled();
});