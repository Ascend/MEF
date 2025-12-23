import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import Wireless from '@/views/manager/network/Wireless.vue';
import * as network from '@/api/network';


const queryWirelessConfigInfoMock = vi.spyOn(network, 'queryWirelessConfigInfo');
const configWirelessStatusInfoMock = vi.spyOn(network, 'configWirelessStatusInfo');
const queryWirelessStatusInfoMock = vi.spyOn(network, 'queryWirelessStatusInfo');
const configWirelessMock = vi.spyOn(network, 'configWireless');

queryWirelessConfigInfoMock.mockResolvedValue({ data: 'data' });
configWirelessStatusInfoMock.mockResolvedValue({ data: 'mock-data' });
queryWirelessStatusInfoMock.mockResolvedValue({ data: 'mock-data' });
configWirelessMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Wireless, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('should test confirmLteDialog', async () => {
  await wrapper.vm.confirmLteDialog();

  expect(configWirelessStatusInfoMock).toHaveBeenCalledWith({
    'state_data': false,
    'state_lte': true,
  });

  expect(wrapper.vm.lteDialog.isShow).toBe(false);
});