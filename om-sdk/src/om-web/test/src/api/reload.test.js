import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { resetSystem } from '@/api/reload';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $postMock = vi.spyOn(httpModule, '$post');

it('resetSystem test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await resetSystem(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Actions/ComputerSystem.Reset';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});
