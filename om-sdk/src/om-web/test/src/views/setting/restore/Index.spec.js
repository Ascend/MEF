import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus, * as elementPlus from 'element-plus';
import i18n from '@/utils/locale';
import Index from '@/views/setting/restore/Index.vue';
import * as reset from '@/api/reset';
import * as commonMethods from '@/utils/commonMethods';


const restoreDefaultsMock = vi.spyOn(reset, 'restoreDefaults');
const confirmMock = vi.spyOn(elementPlus.ElMessageBox, 'confirm');
const clearSessionAndJumpToLoginMock = vi.spyOn(commonMethods, 'clearSessionAndJumpToLogin');

restoreDefaultsMock.mockResolvedValue({ data: 'mock-data' });
confirmMock.mockResolvedValue({ data: 'mock-data' });
clearSessionAndJumpToLoginMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test setting restore submitReset', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  let formElementMock = {
    Password: 'mock',
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.resetForm.isRetain = true;
  wrapper.vm.resetForm.Password = 'mock';
  await wrapper.vm.submitReset(formElementMock);

  expect(validateMock).toHaveBeenCalled();
  expect(confirmMock).toHaveBeenCalled();
});

