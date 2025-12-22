import { describe, expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import CurrAlarm from '@/views/manager/alarm/CurrAlarm.vue';
import * as commonMethods from '@/utils/commonMethods';
import * as alarm from '@/api/alarm';
import * as common from '@/api/common';

const getLanguageDefaultChineseMock = vi.spyOn(commonMethods, 'getLanguageDefaultChinese');
const queryAlarmSourceServiceMock = vi.spyOn(alarm, 'queryAlarmSourceService');
const fetchJsonMock = vi.spyOn(common, 'fetchJson');

getLanguageDefaultChineseMock.mockResolvedValue({ data: 'mock-data' });
const tableDataMock = [
  {
    'AlarmId': '00000002',
    'AlarmName': 'Drive Not Found',
    'AlarmInstance': 'HARD_DISK0',
    'Timestamp': '2023-06-24 16:01:23',
    'PerceivedSeverity': '2',
  },
];
queryAlarmSourceServiceMock.mockResolvedValue({
  data: {
    'AlarMessages': tableDataMock,
  },
});
fetchJsonMock.mockResolvedValue({
  EventSuggestion: [],
});

const wrapper = shallowMount(CurrAlarm, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager alarm CurrAlarm initAlarmTable', async () => {
  const isShowLoading = true;
  const AutoRefresh = false;
  const { initAlarmTable } = wrapper.vm;
  await initAlarmTable(isShowLoading, AutoRefresh);

  expect(wrapper.vm.originAlarmData).toStrictEqual(tableDataMock);
  expect(wrapper.vm.filterData).toStrictEqual(tableDataMock);
  expect(wrapper.vm.alarmData).toStrictEqual(tableDataMock);
});

describe('test manager alarm CurrAlarm searchAlarmTable', () => {
  it('by AlarmId', async () => {
    wrapper.vm.alarmQuery = '00000002';
    const { searchAlarmTable } = wrapper.vm;
    await searchAlarmTable();

    expect(wrapper.vm.filterData).toStrictEqual(tableDataMock);
    expect(wrapper.vm.alarmData).toStrictEqual(tableDataMock);
  });

  it('by AlarmInstance', async () => {
    wrapper.vm.alarmQuery = 'HARD_DISK0';
    const { searchAlarmTable } = wrapper.vm;
    await searchAlarmTable();

    expect(wrapper.vm.filterData).toStrictEqual(tableDataMock);
    expect(wrapper.vm.alarmData).toStrictEqual(tableDataMock);
  });

  it('by AlarmName', async () => {
    wrapper.vm.alarmQuery = 'Drive Not Found';
    const { searchAlarmTable } = wrapper.vm;
    await searchAlarmTable();

    expect(wrapper.vm.filterData).toStrictEqual(tableDataMock);
    expect(wrapper.vm.alarmData).toStrictEqual(tableDataMock);
  });
});

it('test manager alarm CurrAlarm cancelStop', async () => {
  const { cancelStop } = wrapper.vm;
  await cancelStop();

  expect(wrapper.vm.isShowDetailDialog).toBe(false);
});

it('test manager alarm CurrAlarm confirmStop', async () => {
  const { confirmStop } = wrapper.vm;
  await confirmStop();

  expect(wrapper.vm.selectedAlarmItem.isShowDetailDrawer).toBe(true);
  expect(wrapper.vm.isStartRefresh).toBe(false);
  expect(wrapper.vm.isShowDetailDialog).toBe(false);
});

it('test manager alarm CurrAlarm filterPerceivedSeverity', async () => {
  const { filterPerceivedSeverity } = wrapper.vm;
  const row = {
    PerceivedSeverity: 10,
  };
  const rest = filterPerceivedSeverity(10, row);
  expect(rest).toBe(true);
});