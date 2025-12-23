import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/information/Index.vue';
import * as information from '@/api/information';

const queryAllExtendedDevicesInfoMock = vi.spyOn(information, 'queryAllExtendedDevicesInfo');
const querySystemsSourceInfoMock = vi.spyOn(information, 'querySystemsSourceInfo');
const modifySystemsSourceInfoMock = vi.spyOn(information, 'modifySystemsSourceInfo');

queryAllExtendedDevicesInfoMock.mockResolvedValue({ data: { Members: [] } });
querySystemsSourceInfoMock.mockResolvedValue({ data: 'mock-data' });
modifySystemsSourceInfoMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager information handleCurrentChange', async () => {
  const mockValue = 3;
  wrapper.vm.originDeviceData = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9];
  wrapper.vm.pagination.pageSize = 2;
  wrapper.vm.handleCurrentChange(mockValue);
  expect(wrapper.vm.pagination.pageNum).toBe(mockValue);
  expect(wrapper.vm.deviceData).toStrictEqual([4, 5]);
});