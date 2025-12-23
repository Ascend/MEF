import { describe, expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus, * as elementplus from 'element-plus';
import i18n from '@/utils/locale';
import Wired from '@/views/manager/network/Wired.vue';
import * as commonMethods from '@/utils/commonMethods';
import * as common from '@/api/common';
import * as network from '@/api/network';

const warningMock = vi.spyOn(commonMethods.AppMessage, 'warning');
const confirmMock = vi.spyOn(elementplus.ElMessageBox, 'confirm');

const modifyByOdataUrlMock = vi.spyOn(common, 'modifyByOdataUrl');
const queryByOdataUrlMock = vi.spyOn(common, 'queryByOdataUrl');

const queryAllEthernetInfoMock = vi.spyOn(network, 'queryAllEthernetInfo');
const querySingleEthernetInfoMock = vi.spyOn(network, 'querySingleEthernetInfo');

warningMock.mockResolvedValue({ data: 'mock-data' });
confirmMock.mockResolvedValue({ data: 'mock-data' });

modifyByOdataUrlMock.mockResolvedValue({
  data: {
    Oem: {
      TaskPercentage: 'ok',
    },
  },
});
queryByOdataUrlMock.mockResolvedValue({ data: 'mock-data' });

queryAllEthernetInfoMock.mockResolvedValue({ data: 'mock-data' });
querySingleEthernetInfoMock.mockResolvedValue({
  data: {
    Oem: {
      WorkMode: 'mockData',
    },
  },
});

const wrapper = shallowMount(Wired, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

describe('should test manager network Wired openAddIpDialog', () => {
  const { openAddIpDialog } = wrapper.vm;
  it('eth.ethData.length equal MAX_ETH_NUM', () => {
    const ethMock = {
      ethData: {
        key0: '',
        key1: '',
        key2: '',
        key3: '',
      },
    };
    openAddIpDialog(ethMock);
    expect(warningMock).toHaveBeenCalledWith('当前网卡IP地址已达到数量限制');
  });

  it('eth.ethData.length not equal MAX_ETH_NUM', () => {
    const ethMock = {
      ethData: {
        key0: '',
      },
    };
    openAddIpDialog(ethMock);
    expect(wrapper.vm.selectedEthItem.isShow).toBe(true);
    expect(wrapper.vm.selectedEthItem.operateType).toBe('add');
    expect(wrapper.vm.selectedEthItem.currEth).toStrictEqual(ethMock);
    expect(wrapper.vm.selectedEthItem.currIpIndex).toBe(null);
    expect(wrapper.vm.selectedEthItem.wiredData).toBe(wrapper.vm.wiredData);
  });
});

it('should test manager network Wired modifyEthIp', () => {
  const { modifyEthIp } = wrapper.vm;
  const ethMock = {
    ethData: {
      key0: '',
    },
  };
  modifyEthIp(ethMock, 0);
  expect(wrapper.vm.selectedEthItem.isShow).toBe(true);
  expect(wrapper.vm.selectedEthItem.operateType).toBe('edit');
  expect(wrapper.vm.selectedEthItem.currEth).toStrictEqual(ethMock);
  expect(wrapper.vm.selectedEthItem.currIpIndex).toBe(0);
  expect(wrapper.vm.selectedEthItem.wiredData).toBe(wrapper.vm.wiredData);
});

describe('should test manager network Wired deleteEthIp', () => {
  const { deleteEthIp } = wrapper.vm;
  it('eth.ethData.length equal MIN_ETH_NUM', () => {
    const ethMock = {
      ethData: {
        key0: '',
      },
    };
    deleteEthIp(ethMock, 0);
    expect(warningMock).toHaveBeenCalledWith('IP地址不能为空');
  });

  it('eth.ethData.length not equal MIN_ETH_NUM', () => {
    const ethMock = {
      ethData: {
        0: {
          Address: 'mockAddress',
        },
        1: '',
      },
    };
    deleteEthIp(ethMock, 0);
    expect(wrapper.vm.selectedDeleteEthItem.deleteTip).toBe('“mockAddress” IP地址删除后将无法找回，确定删除吗?');
    expect(wrapper.vm.selectedDeleteEthItem.isShowDeleteDialog).toBe(true);
    expect(wrapper.vm.selectedDeleteEthItem.currEth).toStrictEqual(ethMock);
    expect(wrapper.vm.selectedDeleteEthItem.currIpIndex).toBe(0);
  });
});

it('should test manager network Wired clickTestNetwork', async () => {
  const { clickTestNetwork } = wrapper.vm;
  const ethMock = {
    id: 0,
    workMode: '',
  };
  await clickTestNetwork(ethMock);
  expect(ethMock.workMode).toBe('mockData');
});

it('should test manager network Wired cancelAddIp', async () => {
  await wrapper.vm.cancelAddIp();
  expect(wrapper.vm.selectedEthItem.isShow).toBe(false);
});

describe('should test manager network Wired confirmDelete', () => {
  it('selectedDeleteEthItem.value.currEth.isEth0 is true', async () => {
    wrapper.vm.selectedDeleteEthItem = {
      currIpIndex: 0,
      currEth: {
        isEth0: true,
        ethData: ['mockEthData1', 'mockEthData2'],
      },
    };
    await wrapper.vm.confirmDelete();
    expect(wrapper.vm.wiredData.eth0.ethData).toStrictEqual(['mockEthData2']);

    expect(wrapper.vm.selectedDeleteEthItem.deleteTip).toBe('');
    expect(wrapper.vm.selectedDeleteEthItem.isShowDeleteDialog).toBe(false);
    expect(wrapper.vm.selectedDeleteEthItem.currEth).toStrictEqual(null);
    expect(wrapper.vm.selectedDeleteEthItem.currIpIndex).toBe(null);
  });

  it('selectedDeleteEthItem.value.currEth.isEth0 is false', async () => {
    wrapper.vm.selectedDeleteEthItem = {
      currIpIndex: 0,
      currEth: {
        isEth0: false,
        ethData: ['mockEthData1', 'mockEthData2'],
      },
    };
    await wrapper.vm.confirmDelete();
    expect(wrapper.vm.wiredData.eth1.ethData).toStrictEqual(['mockEthData2']);
  });
});

it('should test manager network Wired cancelDelete', () => {
  wrapper.vm.cancelDelete();
  expect(wrapper.vm.selectedDeleteEthItem.deleteTip).toBe('');
  expect(wrapper.vm.selectedDeleteEthItem.isShowDeleteDialog).toBe(false);
  expect(wrapper.vm.selectedDeleteEthItem.currEth).toStrictEqual(null);
  expect(wrapper.vm.selectedDeleteEthItem.currIpIndex).toBe(null);
});

it('should test manager network Wired clickSave', () => {
  const ethMock = {
    odata: '',
    ethData: '',
  };
  wrapper.vm.clickSave(ethMock);
  expect(confirmMock).toHaveBeenCalled();
  expect(wrapper.vm.loading).toBe(false);

  expect(wrapper.vm.selectedDeleteEthItem.deleteTip).toBe('');
  expect(wrapper.vm.selectedDeleteEthItem.isShowDeleteDialog).toBe(false);
  expect(wrapper.vm.selectedDeleteEthItem.currEth).toStrictEqual(null);
  expect(wrapper.vm.selectedDeleteEthItem.currIpIndex).toBe(null);
});
