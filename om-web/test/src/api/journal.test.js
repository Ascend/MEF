import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { downloadLogInfo, queryLogCollectProgress, queryLogsInfo } from '@/api/journal';
import constants from '@/utils/constants';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');

it('queryLogsInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryLogsInfo();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LogServices';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('downloadLogInfo test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await downloadLogInfo(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LogServices/Actions/download';
  expect($postMock).toHaveBeenCalledWith(url, params, { timeout: 5 * constants.MINUTE_TIMEOUT, responseType: 'blob' });

  $postMock.mockRestore();
});

it('queryLogCollectProgress test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryLogCollectProgress();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/LogServices/progress';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});