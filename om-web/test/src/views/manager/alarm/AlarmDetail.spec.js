import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import AlarmDetail from '@/views/manager/alarm/AlarmDetail.vue';
import * as commonMethods from '@/utils/commonMethods';
import * as common from '@/api/common';

const getLanguageDefaultChineseMock = vi.spyOn(commonMethods, 'getLanguageDefaultChinese');
getLanguageDefaultChineseMock.mockResolvedValue({ data: 'mock-data' });
const fetchJsonMock = vi.spyOn(common, 'fetchJson');
fetchJsonMock.mockResolvedValue({
  EventSuggestion: [],
});

const wrapper = shallowMount(AlarmDetail, {
  global: {
    plugins: [ElementPlus, i18n],
  },
  props: {
    alarmItem: {
      isShowDetailDrawer: true,
    },
  },
});

it('test manager alarm AlarmDetail closeDrawer', async () => {
  const { closeDrawer } = wrapper.vm;
  await closeDrawer();
  expect(wrapper.emitted('closeDetailDrawer')).toBeTruthy();
});

