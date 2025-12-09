import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/registration/Index.vue';
import * as registration from '@/api/registration';


const queryNetManagerInfoMock = vi.spyOn(registration, 'queryNetManagerInfo');
const queryNetManagerNodeIDMock = vi.spyOn(registration, 'queryNetManagerNodeID');
const queryFdRootCertMock = vi.spyOn(registration, 'queryFdRootCert');
const modifyNetManagerInfoMock = vi.spyOn(registration, 'modifyNetManagerInfo');

const NetManagerConfigMock = {
  data: {
    'Id': 'NetManager',
    'Name': 'NetManager',
    'NetManager': 'Web',
    'NetIP': '',
    'Port': '',
    'NetAccount': '',
    'ServerName': '',
    'ConnectStatus': 'not_configured',
    'NodeID': {
      '@odata.id': '/redfish/v1/NetManager/NodeID',
    },
    'QueryFdCert': {
      '@odata.id': '/redfish/v1/NetManager/QueryFdCert',
    },
    'Actions': {
      '#ImportFdCert': {
        'target': '/redfish/v1/NetManager/ImportFdCert',
      },
      '#ImportFdCrl': {
        'target': '/redfish/v1/NetManager/ImportFdCrl',
      },
    },
  },
};

queryNetManagerInfoMock.mockResolvedValue(NetManagerConfigMock);
queryNetManagerNodeIDMock.mockResolvedValue({ data: 'mock-data' });
queryFdRootCertMock.mockResolvedValue({ data: 'mock-data' });
modifyNetManagerInfoMock.mockResolvedValue({ data: 'mock-data' });
const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager registration isWebMode', async () => {
  const isWebModeMock = wrapper.vm.isWebMode(NetManagerConfigMock.data.NetManager);
  expect(isWebModeMock).toBe(true)
});