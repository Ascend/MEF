import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import {
  deletePunyDict,
  exportLoginRules,
  exportPunyDict,
  importLoginRules,
  importPunyDict,
  modifyHttpsCertAlarmTime,
  modifyLoginRules,
  queryHttpsCertAlarmTime,
  queryLoginRules
} from '@/api/safety';
import constants from '@/utils/constants';

const SECURITYLOAD_URL = '/redfish/v1/Systems/SecurityService/SecurityLoad';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');
const $patchMock = vi.spyOn(httpModule, '$patch');


it('queryLoginRules test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryLoginRules();
  expect(result).toEqual(mockResponse);
  const url = SECURITYLOAD_URL;
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('modifyLoginRules test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await modifyLoginRules(params);
  expect(result).toEqual(mockResponse);
  const url = SECURITYLOAD_URL;
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('importLoginRules test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await importLoginRules(params);
  expect(result).toEqual(mockResponse);
  const url = `${SECURITYLOAD_URL}/Actions/SecurityLoad.Import`;
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('exportLoginRules test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await exportLoginRules(params);
  expect(result).toEqual(mockResponse);
  const url = `${SECURITYLOAD_URL}/Actions/SecurityLoad.Export`;
  expect($postMock).toHaveBeenCalledWith(url, params, {
    timeout: 1 * constants.MINUTE_TIMEOUT,
  });

  $postMock.mockRestore();
});

it('importPunyDict test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await importPunyDict(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictImport';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('exportPunyDict test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await exportPunyDict(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictExport';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('deletePunyDict test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await deletePunyDict(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictDelete';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('queryHttpsCertAlarmTime test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryHttpsCertAlarmTime();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/HttpsCertAlarmTime';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('modifyHttpsCertAlarmTime test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await modifyHttpsCertAlarmTime(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SecurityService/HttpsCertAlarmTime';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});