import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus, * as elementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/journal/Index.vue';

import * as journal from '@/api/journal';
import * as information from '@/api/information';

import * as commonMethods from '@/utils/commonMethods';
import * as downloadFile from '@/utils/downloadFile';

const queryLogsInfoMock = vi.spyOn(journal, 'queryLogsInfo');
const downloadLogInfoMock = vi.spyOn(journal, 'downloadLogInfo');
const queryLogCollectProgressMock = vi.spyOn(journal, 'queryLogCollectProgress');
const querySystemsSourceInfoMock = vi.spyOn(information, 'querySystemsSourceInfo');

const clearSessionAndJumpToLoginMock = vi.spyOn(commonMethods, 'clearSessionAndJumpToLogin');
const saveFileMock = vi.spyOn(downloadFile, 'saveFile');
const confirmMock = vi.spyOn(elementPlus.ElMessageBox, 'confirm');

queryLogsInfoMock.mockResolvedValue({
  data: {
    'Name': 'mockName',
    'Members@odata.count': 5,
    'Members': [
      {
        '@odata.id': '/redfish/v1/Systems/LogServices/NPU',
      },
      {
        '@odata.id': '/redfish/v1/Systems/LogServices/MindXOM',
      },
      {
        '@odata.id': '/redfish/v1/Systems/LogServices/MEF',
      },
      {
        '@odata.id': '/redfish/v1/Systems/LogServices/OSDivers',
      },
      {
        '@odata.id': '/redfish/v1/Systems/LogServices/MCU',
      },
    ],
  },
});
downloadLogInfoMock.mockResolvedValue({ data: 'mock-data' });
queryLogCollectProgressMock.mockResolvedValue({ data: 'mock-data' });
querySystemsSourceInfoMock.mockResolvedValue({
  data: {
    HostName: 'mockName',
    Model: 'mockModel',
  },
});

clearSessionAndJumpToLoginMock.mockResolvedValue({ data: 'mock-data' });
saveFileMock.mockResolvedValue({ data: 'mock-data' });
confirmMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager journal startCollect', async () => {
  await wrapper.vm.startCollect();
  expect(confirmMock).toHaveBeenCalled();
  expect(querySystemsSourceInfoMock).toHaveBeenCalled();
  expect(queryLogsInfoMock).toHaveBeenCalled();

  const params = { name: 'NPU MindXOM MEF OSDivers MCU' };
  expect(downloadLogInfoMock).toHaveBeenCalledWith(params);
  expect(saveFileMock).toHaveBeenCalled();

  expect(wrapper.vm.isStartCollect).toBe(false);
});

