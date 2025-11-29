import { expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import * as common from '@/api/common';
import * as alarm from '@/api/alarm';
import * as information from '@/api/information';
import * as commonMethods from '@/utils/commonMethods';

import { createRouter, createWebHistory } from 'vue-router';

import HomeView from '@/views/HomeView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/blank',
      component: () => import('@/views/Login.vue'),
    },
    {
      path: '/manager/information',
      component: () => import('@/views/Login.vue'),
    },
  ],
});

const fetchConfigJsonMock = vi.spyOn(common, 'fetchJson');
const queryByOdataUrlMock = vi.spyOn(common, 'queryByOdataUrl');

const queryAlarmSourceServiceMock = vi.spyOn(alarm, 'queryAlarmSourceService');

const querySystemsSourceInfoMock = vi.spyOn(information, 'querySystemsSourceInfo');
const queryCpuInfoMock = vi.spyOn(information, 'queryCpuInfo');
const queryMemoryInfoMock = vi.spyOn(information, 'queryMemoryInfo');
const queryAiProcessorInfoMock = vi.spyOn(information, 'queryAiProcessorInfo');
const queryAllExtendedDevicesInfoMock = vi.spyOn(information, 'queryAllExtendedDevicesInfo');

const getLanguageDefaultChineseMock = vi.spyOn(commonMethods, 'getLanguageDefaultChinese');
const checkModuleTypeMock = vi.spyOn(commonMethods, 'checkModuleType');

const mockResponse = {
  'model': 'Atlas 500 A2',
  'systemName': {
    'zh': 'mock data',
    'en': 'mock data',
  },
  'websiteTitle': {
    'zh': 'mock data',
    'en': 'mock data',
  },
  'userGuide': {
    'zh': 'mock data',
    'en': 'mock data',
  },
  'copyRight': {
    'zh': 'mock data',
    'en': 'mock data',
  },
};

fetchConfigJsonMock.mockResolvedValue(mockResponse);
queryByOdataUrlMock.mockResolvedValue({ data: 'mock-data' });

queryAlarmSourceServiceMock.mockResolvedValue({ data: 'mock-data' });

querySystemsSourceInfoMock.mockResolvedValue({ data: 'mock-data' });
queryCpuInfoMock.mockResolvedValue({ data: 'mock-data' });
queryMemoryInfoMock.mockResolvedValue({ data: 'mock-data' });
queryAiProcessorInfoMock.mockResolvedValue({ data: 'mock-data' });
queryAllExtendedDevicesInfoMock.mockResolvedValue({ data: 'mock-data' });

getLanguageDefaultChineseMock.mockResolvedValue({ data: 'mock-data' });
checkModuleTypeMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(HomeView, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

fetchConfigJsonMock.mockRestore();

it('should test HomeView clickToInformation', async () => {
  wrapper.vm.clickToInformation();

  await router.isReady();

  expect(router.currentRoute.value.path).toBe('/manager/information');
});
