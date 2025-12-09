import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import {
  modifySystemsSourceInfo,
  queryAiProcessorInfo,
  queryAllExtendedDevicesInfo,
  queryCpuInfo,
  queryMemoryInfo,
  querySystemsSourceInfo
} from '@/api/information';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $patchMock = vi.spyOn(httpModule, '$patch');
const isShowLoading = true;
const AutoRefresh = false;

it('queryAllExtendedDevicesInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryAllExtendedDevicesInfo(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/ExtendedDevices';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });

  $getMock.mockRestore();
});

it('modifySystemsSourceInfo test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await modifySystemsSourceInfo(params, isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems';
  expect($patchMock).toHaveBeenCalledWith(url, params, {
    customParams: { isShowLoading },
    timeout: 3 * 60 * 1000,
  });

  $patchMock.mockRestore();
});

it('querySystemsSourceInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await querySystemsSourceInfo(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });

  $getMock.mockRestore();
});

it('queryCpuInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryCpuInfo(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Processors/CPU';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });

  $getMock.mockRestore();
});

it('queryMemoryInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryMemoryInfo(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Memory';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });

  $getMock.mockRestore();
});

it('queryAiProcessorInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryAiProcessorInfo(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Processors/AiProcessor';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });

  $getMock.mockRestore();
});
