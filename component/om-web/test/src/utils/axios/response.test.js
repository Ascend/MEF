import { it, expect } from 'vitest';
import handleResponse from '@/utils/axios/response';


it('returns the response for successful status codes', () => {
  const mockResponse = {
    status: 200,
    data: 'mock data',
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual(mockResponse);
});

it('returns an error object for unsuccessful status codes 400', () => {
  const mockResponse = {
    status: 400,
    data: { code: 400 },
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual({
    code: 400,
    message: '请求错误',
  });
});

it('returns an error object for unsuccessful status codes 401', () => {
  const mockResponse = {
    status: 401,
    data: { code: 401 },
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual({
    code: 401,
    message: '未授权',
  });
});

it('returns an error object for unsuccessful status codes 403', () => {
  const mockResponse = {
    status: 403,
    data: { code: 403 },
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual({
    code: 403,
    message: '未登录',
  });
});

it('returns an error object for unsuccessful status codes 404', () => {
  const mockResponse = {
    status: 404,
    data: { code: 404 },
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual({
    code: 404,
    message: '请求地址错误',
  });
});

it('returns an error object for unsuccessful status codes 412', () => {
  const mockResponse = {
    status: 412,
    data: { code: 412 },
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual({
    code: 412,
    message: '未找到有效session',
  });
});

it('returns an error object for unsuccessful status codes default', () => {
  const mockResponse = {
    status: 500,
    data: { code: 500 },
  };
  const result = handleResponse(mockResponse);
  expect(result).toEqual({
    code: 500,
    message: undefined,
  });
});

