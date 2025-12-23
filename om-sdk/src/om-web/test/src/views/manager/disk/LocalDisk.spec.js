import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import LocalDisk from '@/views/manager/disk/LocalDisk.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const wrapper = mount(LocalDisk, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test manager disk LocalDisk createPartitionSuccessfully', () => {
  wrapper.vm.cancelCreatePartition = vi.fn();
  wrapper.vm.refreshWholeTable = vi.fn();
  wrapper.vm.createPartitionSuccessfully();
});