import { expect, it, vi } from 'vitest';
import * as accountModule from '@/api/account';
import * as httpModule from '@/api/http';

const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');
const $deleteMock = vi.spyOn(httpModule, '$delete');
const $patchMock = vi.spyOn(httpModule, '$patch');

it('createSession test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await accountModule.createSession(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/SessionService/Sessions';
  expect($postMock).toHaveBeenCalledWith(url, params);

  $postMock.mockRestore();
});

it('deleteSession test', async () => {
  $deleteMock.mockResolvedValue(mockResponse);

  const sessionId = 'test-session-id';
  const result = await accountModule.deleteSession(sessionId);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/SessionService/Sessions/' + sessionId;
  expect($deleteMock).toHaveBeenCalledWith(url);

  $deleteMock.mockRestore();
});

it('getUserInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const userId = 'test-user-id';
  const result = await accountModule.getUserInfo(userId);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/AccountService/Accounts/' + userId;
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('modifyUserInfo test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const userId = 'test-user-id';
  const result = await accountModule.modifyUserInfo(userId, params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/AccountService/Accounts/' + userId;
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('querySessionService test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const isShowLoading = true;
  const AutoRefresh = false;
  const result = await accountModule.querySessionService(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/SessionService';
  expect($getMock).toHaveBeenCalledWith(url, {}, { customParams: { isShowLoading }, headers: { AutoRefresh } });

  $getMock.mockRestore();
});

it('modifySessionService test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await accountModule.modifySessionService(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/SessionService';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('queryAccountService test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await accountModule.queryAccountService();
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/AccountService';
  expect($getMock).toHaveBeenCalledWith(url);

  $getMock.mockRestore();
});

it('modifyAccountService test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await accountModule.modifyAccountService(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/AccountService';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});