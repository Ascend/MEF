import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as certificate from '@/api/certificate';
import * as downloadFile from '@/utils/downloadFile';
import CertInfo from '@/views/setting/safety/CertInfo.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const downloadCSRFileMock = vi.spyOn(certificate, 'downloadCSRFile');
const queryHttpsCertInfoMock = vi.spyOn(certificate, 'queryHttpsCertInfo');
const saveFileMock = vi.spyOn(downloadFile, 'saveFile');
downloadCSRFileMock.mockResolvedValue({ data: 'mock-data' });
queryHttpsCertInfoMock.mockResolvedValue({ data: 'mock-data' });
saveFileMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(CertInfo, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety CertInfo downloadCSR', async () => {
  await wrapper.vm.downloadCSR();
  const isDownloadCSRMock = wrapper.vm.isDownloadCSR;
  expect(isDownloadCSRMock).toBe(true);
  expect(downloadCSRFileMock).toHaveBeenCalled();
  expect(saveFileMock).toHaveBeenCalled();
});
