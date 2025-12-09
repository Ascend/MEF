import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';
import ModifyEthIp from '@/views/manager/network/ModifyEthIp.vue';
import * as commonMethods from '@/utils/commonMethods';
import { getLanguageDefaultChinese } from '@/utils/commonMethods';
import * as common from '@/api/common';

const getLanguageDefaultChineseMock = vi.spyOn(commonMethods, 'getLanguageDefaultChinese');
const showWarningAlertMock = vi.spyOn(commonMethods, 'showWarningAlert');
const fetchConfigJsonMock = vi.spyOn(common, 'fetchJson');
getLanguageDefaultChineseMock.mockResolvedValue({ data: 'mock-data' });
showWarningAlertMock.mockResolvedValue({ data: 'mock-data' });
fetchConfigJsonMock.mockResolvedValue({ data: 'mock-data' });

const ethDataMock = {
  Tag: 'web',
  Address: 'mockAddress',
  SubnetMask: 'mockSubnetMask',
  Gateway: '',
  VlanId: null,
  ConnectTest: true,
  RemoteTestIp: '',
  AddressOrigin: 'Static',
};

const wrapper = shallowMount(ModifyEthIp, {
  global: {
    plugins: [ElementPlus, i18n],
  },
  props: {
    selectedEthItem: {
      currEth: {
        isEth0: true,
        ethData: [ethDataMock],
      },
      wiredData: {
        eth1: {
          ethData: [ethDataMock],
        },
        eth0: '',
      },
    },
  },
});

it('test manager network ModifyEthIp handleConfirmAddIp', async () => {
  const { handleConfirmAddIp } = wrapper.vm;
  wrapper.vm.addIpForm = {
    Tag: 'web',
    Address: 'mockAddress',
    SubnetMask: 'mockSubnetMask',
    Gateway: '',
    VlanId: null,
    ConnectTest: true,
    RemoteTestIp: '',
    AddressOrigin: 'Static',
  };
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  let addIpFormRefMock = {
    validate: validateMock,
  };
  wrapper.vm.addIpFormRef = addIpFormRefMock;

  await handleConfirmAddIp();
  expect(showWarningAlertMock).toHaveBeenCalled();
  expect(wrapper.emitted('refresh')).toBeTruthy();
  expect(wrapper.emitted('temperatureSaveIpAddresses')).toBeTruthy();
});

it('test manager network ModifyEthIp handleCancelAddIp', async () => {
  const validateMock = vi.fn((callback) => {
    callback(true);
  });
  let addIpFormRefMock = {
    validate: validateMock,
    resetFields: vi.fn(),
  };
  wrapper.vm.addIpFormRef = addIpFormRefMock;
  await wrapper.vm.handleCancelAddIp();
  expect(addIpFormRefMock.resetFields).toHaveBeenCalled();
  expect(wrapper.emitted('cancelAddIp')).toBeTruthy();
});

