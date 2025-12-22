import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import { downloadCSRFile, importServerCertificate, queryHttpsCertInfo } from '@/api/certificate';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');

it('queryHttpsCertInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryHttpsCertInfo();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/HttpsCert';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('importServerCertificate test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await importServerCertificate(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/HttpsCert/Actions/HttpsCert.ImportServerCertificate';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('downloadCSRFile test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await downloadCSRFile(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/downloadCSRFile';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});
