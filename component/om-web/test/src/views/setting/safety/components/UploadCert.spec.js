import { it, vi, expect } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import { createRouter, createWebHistory } from 'vue-router';

import * as commonMethods from '@/utils/commonMethods';
import * as certificate from '@/api/certificate';

import UploadCert from '@/views/setting/safety/components/UploadCert.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

const isValidUploadFilenameMock = vi.spyOn(commonMethods, 'isValidUploadFilename');
const handleUploadErrorMock = vi.spyOn(commonMethods, 'handleUploadError');
const importServerCertificateMock = vi.spyOn(certificate, 'importServerCertificate');
handleUploadErrorMock.mockResolvedValue({ data: 'mock-data' });
importServerCertificateMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(UploadCert, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test setting safety components UploadCert beforeUploadFile', async () => {
  const rawFileMock = {
    name: 'nameMock',
    size: 32,
  };
  wrapper.vm.uploadFormRef = {
    resetFields: vi.fn(),
  };
  await wrapper.vm.beforeUploadFile(rawFileMock);
  expect(isValidUploadFilenameMock).toHaveBeenCalledWith(rawFileMock.name);
});

it('test setting safety components UploadCert confirmUpload', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  const formElement = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.upload = {
    submit: vi.fn(),
  };
  await wrapper.vm.confirmUpload(formElement);
  expect(validateMock).toHaveBeenCalled();
});

it('test setting safety components UploadCert uploadSuccess', async () => {
  wrapper.vm.upload = {
    clearFiles: vi.fn(),
  }
  wrapper.vm.uploadFormRef = {
    resetFields: vi.fn(),
  };
  await wrapper.vm.uploadSuccess();
  expect(importServerCertificateMock).toHaveBeenCalled();
});

it('test setting safety components UploadCert uploadError', async () => {
  wrapper.vm.upload = {
    clearFiles: vi.fn(),
  }
  wrapper.vm.uploadFormRef = {
    resetFields: vi.fn(),
  };
  await wrapper.vm.uploadError();
  expect(handleUploadErrorMock).toHaveBeenCalled();
});