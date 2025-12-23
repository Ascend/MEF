import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as drive from '@/api/drive';
import MountPartition from '@/views/manager/disk/MountPartition.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const mountPartitionsMock = vi.spyOn(drive, 'mountPartitions');
const unmountPartitionsMock = vi.spyOn(drive, 'unmountPartitions');
mountPartitionsMock.mockResolvedValue({ data: 'mock-data' });
unmountPartitionsMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(MountPartition, {
  propsData: {
    mountPartitionData: {
      isShow: true,
      partitionId: '12345678',
      location: null,
      diskIndex: null,
    },
  },
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test manager disk MountPartition confirmMountPartition', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.partitionForm.mountPath = '/dev/xxx';
  await wrapper.vm.confirmMountPartition(formElementMock);
  expect(mountPartitionsMock).toHaveBeenCalled();
});

it('test manager disk MountPartition cancelMountPartition', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  await wrapper.vm.cancelMountPartition(formElementMock);
  expect(formElementMock.resetFields).toHaveBeenCalled();
});

it('test manager disk MountPartition confirmUnmountPartition', async () => {
  await wrapper.vm.confirmUnmountPartition();
  expect(unmountPartitionsMock).toHaveBeenCalled();
});