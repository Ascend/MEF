import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as drive from '@/api/drive';
import CreatePartition from '@/views/manager/disk/CreatePartition.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const createPartitionsMock = vi.spyOn(drive, 'createPartitions');
createPartitionsMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(CreatePartition, {
  propsData: {
    createPartitionData: {
      isShow: false,
      currDisk: {
        driveLetter: 'dev',
      },
      diskIndex: null,
    },
  },
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test manager disk CreatePartition confirmCreatePartition', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.partitionForm = {
    capacity: '123456789',
    number: 3,
  }
  await wrapper.vm.confirmCreatePartition(formElementMock);
  expect(createPartitionsMock).toHaveBeenCalled();
});

it('test manager disk CreatePartition cancelCreatePartition', async () => {
  const formElementMock = {
    resetFields: vi.fn(),
  };
  await wrapper.vm.cancelCreatePartition(formElementMock);
  expect(formElementMock.resetFields).toHaveBeenCalled();
});