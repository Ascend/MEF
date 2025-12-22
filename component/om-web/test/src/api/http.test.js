import { describe, expect, it, vi } from 'vitest';
import { $delete, $get, $patch, $post } from '@/api/http';
import { request } from '@/utils/axios';

describe('$get', () => {
  it('sends a GET request and returns the response', async () => {
    const mockResponse = { data: 'mock data' };
    const requestGetMock = vi.spyOn(request, 'get');
    requestGetMock.mockResolvedValue(mockResponse);

    const result = await $get('/test');
    expect(result).toEqual(mockResponse);

    requestGetMock.mockRestore();
  });

  it('handles request errors', async () => {
    const mockError = new Error('mock error');
    const requestPostMock = vi.spyOn(request, 'get');
    requestPostMock.mockRejectedValue(mockError);

    try {
      await $get('/test');
    } catch (err) {
      expect(err).toEqual(mockError);
    }

    requestPostMock.mockRestore();
  });
});

describe('$post', () => {
  it('sends a POST request and returns the response', async () => {
    const mockResponse = { data: 'mock data' };
    const requestPostMock = vi.spyOn(request, 'post');
    requestPostMock.mockResolvedValue(mockResponse);

    const result = await $post('/test');
    expect(result).toEqual(mockResponse);

    requestPostMock.mockRestore();
  });

  it('handles request errors', async () => {
    const mockError = new Error('mock error');
    const requestPostMock = vi.spyOn(request, 'post');
    requestPostMock.mockRejectedValue(mockError);

    try {
      await $post('/test');
    } catch (err) {
      expect(err).toEqual(mockError);
    }

    requestPostMock.mockRestore();
  });
});

describe('$delete', () => {
  it('sends a DELETE request and returns the response', async () => {
    const mockResponse = { data: 'mock data' };
    const requestPostMock = vi.spyOn(request, 'delete');
    requestPostMock.mockResolvedValue(mockResponse);

    const result = await $delete('/test');
    expect(result).toEqual(mockResponse);

    requestPostMock.mockRestore();
  });

  it('handles request errors', async () => {
    const mockError = new Error('mock error');
    const requestPostMock = vi.spyOn(request, 'delete');
    requestPostMock.mockRejectedValue(mockError);

    try {
      await $delete('/test');
    } catch (err) {
      expect(err).toEqual(mockError);
    }

    requestPostMock.mockRestore();
  });
});

describe('$patch', () => {
  it('sends a PATCH request and returns the response', async () => {
    const mockResponse = { data: 'mock data' };
    const requestPostMock = vi.spyOn(request, 'patch');
    requestPostMock.mockResolvedValue(mockResponse);

    const result = await $patch('/test');
    expect(result).toEqual(mockResponse);

    requestPostMock.mockRestore();
  });

  it('handles request errors', async () => {
    const mockError = new Error('mock error');
    const requestPostMock = vi.spyOn(request, 'patch');
    requestPostMock.mockRejectedValue(mockError);

    try {
      await $patch('/test');
    } catch (err) {
      expect(err).toEqual(mockError);
    }

    requestPostMock.mockRestore();
  });
});

