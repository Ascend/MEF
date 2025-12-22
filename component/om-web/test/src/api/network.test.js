import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import {
  configWireless,
  configWirelessStatusInfo,
  queryAllEthernetInfo,
  querySingleEthernetInfo,
  queryWirelessConfigInfo,
  queryWirelessStatusInfo
} from '@/api/network';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $patchMock = vi.spyOn(httpModule, '$patch');
const isShowLoading = true;

it('queryAllEthernetInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryAllEthernetInfo(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/EthernetInterfaces';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
    });

  $getMock.mockRestore();
});

it('querySingleEthernetInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await querySingleEthernetInfo(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/EthernetInterfaces/' + params;
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('queryWirelessStatusInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryWirelessStatusInfo(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LTE/StatusInfo';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
    });

  $getMock.mockRestore();
});

it('configWirelessStatusInfo test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await configWirelessStatusInfo(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LTE/StatusInfo';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('queryWirelessConfigInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryWirelessConfigInfo();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LTE/ConfigInfo';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('configWireless test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await configWireless(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LTE/ConfigInfo';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});