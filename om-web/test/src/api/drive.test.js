import { expect, it, vi } from 'vitest';
import * as httpModule from '@/api/http';
import {
  createPartitions,
  deletePartition,
  mountNfs,
  mountPartitions,
  queryAllPartitionsInfo,
  queryAllSimpleStoragesInfo,
  queryNfsInfo, unmountNfs,
  unmountPartitions
} from '@/api/drive';


const params = { key: 'value' };
const mockResponse = { data: 'mock data' };
const $getMock = vi.spyOn(httpModule, '$get');
const $postMock = vi.spyOn(httpModule, '$post');
const $deleteMock = vi.spyOn(httpModule, '$delete');
const $patchMock = vi.spyOn(httpModule, '$patch');

const isShowLoading = true;
const AutoRefresh = false;

it('queryAllSimpleStoragesInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryAllSimpleStoragesInfo(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/SimpleStorages';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
    });

  $getMock.mockRestore();
});

it('queryAllPartitionsInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryAllPartitionsInfo(isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Partitions';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
    });

  $getMock.mockRestore();
});

it('createPartitions test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await createPartitions(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Partitions';
  expect($postMock).toHaveBeenCalledWith(url, params, {
    timeout: 10 * 60 * 1000,
  });

  $postMock.mockRestore();
});

it('deletePartition test', async () => {
  $deleteMock.mockResolvedValue(mockResponse);

  const partitionId = 'mock-id';
  const result = await deletePartition(partitionId);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Partitions/' + partitionId;
  expect($deleteMock).toHaveBeenCalledWith(url);

  $deleteMock.mockRestore();
});

it('mountPartitions test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await mountPartitions(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Partitions/Mount';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('unmountPartitions test', async () => {
  $patchMock.mockResolvedValue(mockResponse);

  const result = await unmountPartitions(params);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/Partitions/Unmount';
  expect($patchMock).toHaveBeenCalledWith(url, params);

  $patchMock.mockRestore();
});

it('queryNfsInfo test', async () => {
  $getMock.mockResolvedValue(mockResponse);

  const result = await queryNfsInfo(isShowLoading, AutoRefresh);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/NfsManage';
  expect($getMock).toHaveBeenCalledWith(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });

  $getMock.mockRestore();
});

it('mountNfs test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await mountNfs(params, isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/NfsManage/Actions/NfsManage.Mount';
  expect($postMock).toHaveBeenCalledWith(url, params, {
    customParams: { isShowLoading },
  });

  $postMock.mockRestore();
});

it('unmountNfs test', async () => {
  $postMock.mockResolvedValue(mockResponse);

  const result = await unmountNfs(params, isShowLoading);
  expect(result).toEqual(mockResponse);
  const url = '/redfish/v1/Systems/NfsManage/Actions/NfsManage.Unmount';
  expect($postMock).toHaveBeenCalledWith(url, params, {
    customParams: { isShowLoading },
  });

  $postMock.mockRestore();
});