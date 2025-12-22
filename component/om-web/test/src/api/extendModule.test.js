import { it, expect, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { queryAllModules } from '@/api/extendModule';

const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');

it('queryAllModules test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const isShowLoading = true;
  const AutoRefresh = false;
  const url = '/redfish/v1/Systems/Modules';
  const result = await queryAllModules(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  expect($getMock).toHaveBeenCalledWith(url, {}, {
    customParams: { isShowLoading },
    headers: { AutoRefresh },
  });

  $getMock.mockRestore();
});