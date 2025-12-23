import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/reload/Index.vue';

import * as reload from '@/api/reload';
import * as commonMethods from '@/utils/commonMethods';
import { clearSessionAndJumpToLogin } from '@/utils/commonMethods';
import { useRouter } from 'vue-router';

const router = useRouter();
const resetSystemMock = vi.spyOn(reload, 'resetSystem');
const clearSessionAndJumpToLoginMock = vi.spyOn(commonMethods, 'clearSessionAndJumpToLogin');

resetSystemMock.mockResolvedValue({ data: 'mock-data' });
clearSessionAndJumpToLoginMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager reload confirmRestart', async () => {
  await wrapper.vm.confirmRestart();

  expect(wrapper.vm.isConfirmRestart).toBe(false);

  let params = { 'ResetType': 'Restart' };
  expect(resetSystemMock).toHaveBeenCalledWith(params);

  expect(clearSessionAndJumpToLoginMock).toHaveBeenCalledWith(router);
});

it('test manager reload cancelRestart', async () => {
  await wrapper.vm.cancelRestart();

  expect(wrapper.vm.isConfirmRestart).toBe(false);
});

it('test manager reload confirmRestartType graceful', async () => {
  await wrapper.vm.confirmRestartType('graceful');

  expect(wrapper.vm.confirmDialogTitle).toBe('重启提醒');
  expect(wrapper.vm.confirmDialogContent).toBe('确定重启当前系统？');
  expect(wrapper.vm.isConfirmRestart).toBe(true);
});

it('test manager reload confirmRestartType not graceful', async () => {
  await wrapper.vm.confirmRestartType('mock');

  expect(wrapper.vm.confirmDialogTitle).toBe('上下电提醒');
  expect(wrapper.vm.confirmDialogContent).toBe('确定上下电当前系统？');
  expect(wrapper.vm.isConfirmRestart).toBe(true);
});