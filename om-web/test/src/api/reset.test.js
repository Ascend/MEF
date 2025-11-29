import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { queryEthernetIpList, resetAction, restoreDefaults } from '@/api/reset';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');

it('queryEthernetIpList test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryEthernetIpList();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/EthIpList';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('resetAction test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await resetAction(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Actions/RestoreDefaults.Reset';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('restoreDefaults test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await restoreDefaults(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Actions/RestoreDefaults.Config';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});


