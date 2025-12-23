import { it, expect, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { fetchJson, modifyByOdataUrl, queryByOdataUrl } from '@/api/common';
import constants from '@/utils/constants';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $patchMock = vi.spyOn(httpModule, '$patch');

it('queryByOdataUrl test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const isShowLoading = true;
  const AutoRefresh = false;
  const url = 'mockUrl';
  const result = await queryByOdataUrl(url, isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  expect($getMock).toHaveBeenCalledWith(url, {}, {
    customParams: { isShowLoading },
    headers: { AutoRefresh },
  });

  $getMock.mockRestore();
});

it('modifyByOdataUrl test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const url = 'mockUrl';
  const timeout = constants.DEFAULT_TIMEOUT;
  const isShowLoading = true;
  const result = await modifyByOdataUrl(url, params, timeout, isShowLoading);
  expect(result).toEqual(mockResponse);
  expect($patchMock).toHaveBeenCalledWith(url, params, { customParams: { isShowLoading }, timeout });

  $patchMock.mockRestore();
});

it('fetchJson test', async () => {
  $getMock.mockResolvedValue(mockResponse);
  const isShowLoading = false;
  const url = '/WhiteboxConfig/json/config.json';
  const result = await fetchJson(url, isShowLoading);
  expect(result).toEqual('mock data');
  expect($getMock).toHaveBeenCalledWith(url, {}, { customParams: { isShowLoading } });

  $getMock.mockRestore();
});