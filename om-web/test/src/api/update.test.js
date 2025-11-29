import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { queryHddUpgradeInfo, queryUpdateStatus, queryUpgradeFlag, resetFirmware, updateFirmware, updateHdd } from '@/api/update';


const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');

const isShowLoading = true;
const hddId = 'mock-id';

it('queryUpdateStatus test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryUpdateStatus(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
    });

  $getMock.mockRestore();
});

it('updateFirmware test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await updateFirmware(params, isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate';
  expect($postMock).toHaveBeenCalledWith(url, params, {
    customParams: { isShowLoading },
  });

  $postMock.mockRestore();
});

it('resetFirmware test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await resetFirmware(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/UpdateService/Actions/UpdateService.Reset';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('updateHdd test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await updateHdd(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/UpdateHddService/Actions/UpdateHddService.SimpleUpdate';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('queryHddUpgradeInfo test', async () => {
  $postMock.mockResolvedValue(mockResponse);
  let partParams = {
    'HddNo': parseInt(hddId),
  };

  const result = await queryHddUpgradeInfo(hddId);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/UpdateHddService/Actions/UpdateHddService.infos';
  expect($postMock).toHaveBeenCalledWith(url, partParams);

  $postMock.mockRestore();
});

it('queryUpgradeFlag test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryUpgradeFlag();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/UpdateHddService/upgradeFlag';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});