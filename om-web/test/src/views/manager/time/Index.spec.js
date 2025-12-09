import { expect, it, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/time/Index.vue';
import * as time from '@/api/time';
import * as information from '@/api/information';
import { createRouter, createWebHistory } from 'vue-router';


const querySystemTimeMock = vi.spyOn(time, 'querySystemTime');
const queryNTPServiceMock = vi.spyOn(time, 'queryNTPService');
const configNTPServiceMock = vi.spyOn(time, 'configNTPService');

const querySystemsSourceInfoMock = vi.spyOn(information, 'querySystemsSourceInfo');
const modifySystemsSourceInfoMock = vi.spyOn(information, 'modifySystemsSourceInfo');

const dataMock = {
  data: {
    '@odata.context': '/redfish/v1/$metadata#Systems',
    '@odata.id': '/redfish/v1/Systems',
    '@odata.type': '#ComputerSystem.v1_18_0.ComputerSystem',
    'Id': '1',
    'Name': 'Computer System',
    'HostName': 'Euler01',
    'UUID': null,
    'Model': 'Atlas 500 A2',
    'SupportModel': 'Atlas 500 A2',
    'SerialNumber': null,
    'AssetTag': null,
    'Status': {
      'Health': 'OK',
    },
    'Processors': {
      '@odata.id': '/redfish/v1/Systems/Processors',
    },
    'Memory': {
      '@odata.id': '/redfish/v1/Systems/Memory',
    },
    'EthernetInterfaces': {
      '@odata.id': '/redfish/v1/Systems/EthernetInterfaces',
    },
    'LogServices': {
      '@odata.id': '/redfish/v1/Systems/LogServices',
    },
    'SimpleStorages': {
      '@odata.id': '/redfish/v1/Systems/SimpleStorages',
    },
    'Oem': {
      'PCBVersion': 1,
      'Temperature': 35,
      'Power': 13.9,
      'Voltage': 0.02,
      'CpuHeating': null,
      'DiskHeating': 'Stop',
      'UsbHubHeating': null,
      'AiTemperature': 40,
      'SoftwareVersion': '',
      'ProcessorArchitecture': 'ARM',
      'OSVersion': 'EulerOS 2.0 (SP10)',
      'KernelVersion': '4.19.90-vhulk2208.3.0.h1277.eulerosv2r10.aarch64',
      'Uptime': '09:09:40 19 days',
      'Datetime': 'Fri Jul 14 01:09:24 2023',
      'DateTimeLocalOffset': 'Etc/GMT-14 (+14, +1400)',
      'CpuUsage': 0.75,
      'MemoryUsage': 13.55,
      'Firmware': [
        {
          'Version': '22.0.4.b060',
          'InactiveVersion': '',
          'Module': 'NPU Driver',
          'BoardId': '',
          'UpgradeResult': null,
          'UpgradeProcess': 0,
        },
        {
          'Version': '3.3.5',
          'InactiveVersion': '',
          'Module': 'MCU',
          'BoardId': '0x0041',
          'UpgradeResult': null,
          'UpgradeProcess': 0,
        },
        {
          'Version': '',
          'InactiveVersion': '',
          'Module': 'Ascend-firmware',
          'BoardId': '',
          'UpgradeResult': null,
          'UpgradeProcess': 0,
        },
      ],
      'InactiveConfiguration': null,
      'NTPService': {
        '@odata.id': '/redfish/v1/Systems/NTPService',
      },
      'ExtendedDevices': {
        '@odata.id': '/redfish/v1/Systems/ExtendedDevices',
      },
      'LTE': {
        '@odata.id': '/redfish/v1/Systems/LTE',
      },
      'Partitions': {
        '@odata.id': '/redfish/v1/Systems/Partitions',
      },
      'NfsManage': {
        '@odata.id': '/redfish/v1/Systems/NfsManage',
      },
      'SecurityService': {
        '@odata.id': '/redfish/v1/Systems/SecurityService',
      },
      'Alarm': {
        '@odata.id': '/redfish/v1/Systems/Alarm',
      },
      'SystemTime': {
        '@odata.id': '/redfish/v1/Systems/SystemTime',
      },
      'EthIpList': {
        '@odata.id': '/redfish/v1/Systems/EthIpList',
      },
      'DigitalWarranty': {
        '@odata.id': '/redfish/v1/Systems/DigitalWarranty',
      },
    },
    'Actions': {
      '#ComputerSystem.Reset': {
        'target': '/redfish/v1/Systems/Actions/ComputerSystem.Reset',
      },
      'Oem': {
        '#ComputerSystem.RestoreDefaults': {
          'target': '/redfish/v1/Systems/Actions/ComputerSystem.RestoreDefaults',
        },
      },
    },
  },
};

const router = createRouter({
  history: createWebHistory(),
  routes: [{
    path: '/blank',
    component: () => import('@/views/Login.vue'),
  }],
});

querySystemTimeMock.mockResolvedValue({ data: 'mock-data' });
queryNTPServiceMock.mockResolvedValue({ data: 'mock-data' });
configNTPServiceMock.mockResolvedValue({ data: 'mock-data' });

querySystemsSourceInfoMock.mockResolvedValue(dataMock);
modifySystemsSourceInfoMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = mount(Index, {
  global: {
    plugins: [ElementPlus, i18n, router],
  },
});

it('test manager time isEnglish', async () => {
  const isEnglishMock = wrapper.vm.isEnglish();
  expect(isEnglishMock).toBe(false);
});

it('test manager time isEnglish', async () => {
  const dateMock = new Date('Fri Jul 14 01:09:24 2010');
  const setDisableDateRet = wrapper.vm.setDisableDate(dateMock);
  expect(setDisableDateRet).toBe(true);
});