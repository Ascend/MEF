import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import AlarmTableAndDialog from '@/views/manager/alarm/AlarmTableAndDialog.vue';
import * as commonMethods from '@/utils/commonMethods';
import * as alarm from '@/api/alarm';

const getLanguageDefaultChineseMock = vi.spyOn(commonMethods, 'getLanguageDefaultChinese');

const createAlarmShieldRulesMock = vi.spyOn(alarm, 'createAlarmShieldRules');
const cancelAlarmShieldRulesMock = vi.spyOn(alarm, 'cancelAlarmShieldRules');

getLanguageDefaultChineseMock.mockResolvedValue({ data: 'mock-data' });
createAlarmShieldRulesMock.mockResolvedValue({ data: 'mock-data' });
cancelAlarmShieldRulesMock.mockResolvedValue({ data: 'mock-data' });

const tableDataMock = [
  {
    'AlarmId': '00000002',
    'AlarmName': 'persent',
    'AlarmInstance': 'HARD_DISK0',
    'Timestamp': '2023-06-24 16:01:23',
    'PerceivedSeverity': '2',
  },
];

const wrapper = shallowMount(AlarmTableAndDialog, {
  global: {
    plugins: [ElementPlus, i18n],
  },
  props: {
    tableData: tableDataMock,
    tabName: String,
  },
});

it('test manager alarm AlarmTableAndDialog searchAlarmTable', async () => {
  const { searchAlarmTable } = wrapper.vm;
  await searchAlarmTable();
  expect(wrapper.vm.pageNum).toBe(1);
  expect(wrapper.vm.filterData).toStrictEqual(tableDataMock);
  expect(wrapper.vm.alarmData).toStrictEqual(tableDataMock);
});

it('test manager alarm AlarmTableAndDialog refreshAlarmTable', async () => {
  const { refreshAlarmTable } = wrapper.vm;
  await refreshAlarmTable();
  expect(wrapper.emitted('refresh')).toBeTruthy();
});

it('test manager alarm AlarmTableAndDialog handleSizeChange', async () => {
  const { handleSizeChange } = wrapper.vm;
  await handleSizeChange(10);
  expect(wrapper.vm.pageSize).toBe(10);
  expect(wrapper.vm.alarmData).toStrictEqual(tableDataMock);
});

it('test manager alarm AlarmTableAndDialog handleCurrentChange', async () => {
  const { handleCurrentChange } = wrapper.vm;
  await handleCurrentChange(10);
  expect(wrapper.vm.pageNum).toBe(10);
  expect(wrapper.vm.alarmData).toStrictEqual([]);
});

it('test manager alarm AlarmTableAndDialog filterPerceivedSeverity', async () => {
  const { filterPerceivedSeverity } = wrapper.vm;
  const rowMock = {
    PerceivedSeverity: 10,
  };
  const rest = filterPerceivedSeverity(10, rowMock);
  expect(rest).toBe(true);
});

it('test manager alarm AlarmTableAndDialog handleSelectionChange', async () => {
  const { handleSelectionChange } = wrapper.vm;
  handleSelectionChange(10);
  expect(wrapper.vm.selectedAlarm).toBe(10);
});

it('test manager alarm AlarmTableAndDialog clickOperationButton', async () => {
  const { clickOperationButton } = wrapper.vm;
  const rowMock = 'mock';
  clickOperationButton(rowMock);
  expect(wrapper.vm.selectedAlarm).toStrictEqual(['mock']);
  expect(wrapper.vm.isShowDialog).toBe(true);
});

it('test manager alarm AlarmTableAndDialog cancelDialog', async () => {
  const { cancelDialog } = wrapper.vm;
  cancelDialog();
  expect(wrapper.vm.isShowDialog).toBe(false);
});
