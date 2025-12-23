import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import Index from '@/views/manager/alarm/Index.vue';
import * as alarm from '@/api/alarm';
import * as common from '@/api/common';

const queryAlarmShieldRulesMock = vi.spyOn(alarm, 'queryAlarmShieldRules');
const fetchJsonMock = vi.spyOn(common, 'fetchJson');
queryAlarmShieldRulesMock.mockResolvedValue({
  data: {
    AlarmShieldMessages: [],
  },
});
fetchJsonMock.mockResolvedValue({
  EventSuggestion: [],
});

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager alarm Index refresh', async () => {
  await wrapper.vm.refresh();
  expect(fetchJsonMock).toHaveBeenCalled();
  expect(queryAlarmShieldRulesMock).toHaveBeenCalled();
});

