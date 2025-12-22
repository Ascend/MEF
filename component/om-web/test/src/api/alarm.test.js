import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import {
  cancelAlarmShieldRules,
  createAlarmShieldRules,
  queryAlarmShieldRules,
  queryAlarmSourceService
} from '@/api/alarm';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $patchMock = vi.spyOn(httpModule, '$patch');

it('queryAlarmSourceService test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const isShowLoading = true;
  const AutoRefresh = false;
  const result = await queryAlarmSourceService(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Alarm/AlarmInfo';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    { customParams: { isShowLoading }, headers: { AutoRefresh } }
  );

  $getMock.mockRestore();
});

it('queryAlarmShieldRules test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryAlarmShieldRules();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Alarm/AlarmShield';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('createAlarmShieldRules test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await createAlarmShieldRules(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Alarm/AlarmShield/Increase';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('cancelAlarmShieldRules test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await cancelAlarmShieldRules(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Alarm/AlarmShield/Decrease';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});