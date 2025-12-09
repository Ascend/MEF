import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { modifyNetManagerInfo, queryFdRootCert, queryNetManagerInfo, queryNetManagerNodeID } from '@/api/registration';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');

it('queryNetManagerInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryNetManagerInfo();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/NetManager';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('queryNetManagerNodeID test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryNetManagerNodeID();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/NetManager/NodeID';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('queryFdRootCert test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryFdRootCert();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/NetManager/QueryFdCert';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('modifyNetManagerInfo test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await modifyNetManagerInfo(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/NetManager';
  expect($postMock).toHaveBeenCalledWith(url, params, { timeout: 4.25 * 60 * 1000 });

  $postMock.mockRestore();
});