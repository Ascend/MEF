import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as drive from '@/api/drive';
import Nfs from '@/views/manager/disk/Nfs.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const queryNfsInfoMock = vi.spyOn(drive, 'queryNfsInfo');
const mountNfsMock = vi.spyOn(drive, 'mountNfs');
const unmountNfsMock = vi.spyOn(drive, 'unmountNfs');
unmountNfsMock.mockResolvedValue({ data: 'mock-data' });
mountNfsMock.mockResolvedValue({ data: 'mock-data' });
queryNfsInfoMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(Nfs, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test manager disk Nfs confirmCreate', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElementMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.isShowCreate = false;
  wrapper.vm.nfsForm = {
    ServerIP: '',
    ServerDir: '',
    MountPath: '',
    FileSystem: 'nfs4',
  };
  wrapper.vm.nfsFormRef = {
    resetFields: vi.fn(),
  };
  await wrapper.vm.confirmCreate(formElementMock);
  expect(mountNfsMock).toHaveBeenCalled();
});

it('test manager disk Nfs cancelCreate', async () => {
  const formElementMock = {
    resetFields: vi.fn(),
  };
  wrapper.vm.isShowCreate = false;
  await wrapper.vm.cancelCreate(formElementMock);
  expect(formElementMock.resetFields).toHaveBeenCalled();
});