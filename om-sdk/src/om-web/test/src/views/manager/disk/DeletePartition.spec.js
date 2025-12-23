import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as drive from '@/api/drive';
import DeletePartition from '@/views/manager/disk/DeletePartition.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const deletePartitionMock = vi.spyOn(drive, 'deletePartition');
deletePartitionMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(DeletePartition, {
  propsData: {
    deletePartitionData: {
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

it('test manager disk DeletePartition confirmDeletePartition', async () => {
  await wrapper.vm.confirmDeletePartition();
  expect(deletePartitionMock).toHaveBeenCalled();
});
