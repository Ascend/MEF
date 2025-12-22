import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import Index from '@/views/manager/network/Index.vue';
import * as network from '@/api/network';


const queryWirelessStatusInfoMock = vi.spyOn(network, 'queryWirelessStatusInfo');
queryWirelessStatusInfoMock.mockResolvedValue({ data: 'mock-data' });

it('should test manager network Index', async () => {
  const wrapper = shallowMount(Index, {
    global: {
      plugins: [ElementPlus, i18n],
    },
  });
  expect(wrapper.vm.activeName).toBe('wired');
  expect(wrapper.vm.hasWireless).toBe(false);
  expect(queryWirelessStatusInfoMock).toHaveBeenCalled();
});
