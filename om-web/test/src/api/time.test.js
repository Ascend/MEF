import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { configNTPService, queryNTPService, querySystemTime } from '@/api/time';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $patchMock = vi.spyOn(httpModule, '$patch');
const isShowLoading = true;

it('querySystemTime test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await querySystemTime(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SystemTime';
  expect($getMock).toHaveBeenCalledWith(url, {}, { customParams: { isShowLoading } });

  $getMock.mockRestore();
});

it('queryNTPService test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryNTPService(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/NTPService';
  expect($getMock).toHaveBeenCalledWith(url, {}, { customParams: { isShowLoading } });

  $getMock.mockRestore();
});

it('configNTPService test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await configNTPService(params, isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/NTPService';
  expect($patchMock).toHaveBeenCalledWith(url, params, { customParams: { isShowLoading } });

  $patchMock.mockRestore();
});