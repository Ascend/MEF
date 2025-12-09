import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus, * as elementPlus from 'element-plus';
import i18n from '@/utils/locale';
import Index from '@/views/setting/reset/Index.vue';
import * as reset from '@/api/reset';
import * as commonMethods from '@/utils/commonMethods';


const queryEthernetIpListMock = vi.spyOn(reset, 'queryEthernetIpList');
const resetActionMock = vi.spyOn(reset, 'resetAction');

const confirmMock = vi.spyOn(elementPlus.ElMessageBox, 'confirm');
const clearSessionAndJumpToLoginMock = vi.spyOn(commonMethods, 'clearSessionAndJumpToLogin');

queryEthernetIpListMock.mockResolvedValue({ data: 'mock-data' });
resetActionMock.mockResolvedValue({ data: 'mock-data' });
confirmMock.mockResolvedValue({ data: 'mock-data' });
clearSessionAndJumpToLoginMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test setting reset submitReset', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  let formElementMock = {
    Password: 'mock',
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.resetForm.isRetain = true;
  wrapper.vm.resetForm.rootPassword = 'mock';
  await wrapper.vm.submitReset(formElementMock);

  expect(validateMock).toHaveBeenCalled();
  expect(confirmMock).toHaveBeenCalled();
});

