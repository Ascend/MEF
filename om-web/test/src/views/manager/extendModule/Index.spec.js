import { expect, it, vi } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/extendModule/Index.vue';
import * as extendModule from '@/api/extendModule';
import * as common from '@/api/common';


const queryAllModulesMock = vi.spyOn(extendModule, 'queryAllModules');

const queryByOdataUrlMock = vi.spyOn(common, 'queryByOdataUrl');
const modifyByOdataUrlMock = vi.spyOn(common, 'modifyByOdataUrl');

queryAllModulesMock.mockResolvedValue({ data: { Members: [{ '@odata.id': '' }] } });

queryByOdataUrlMock.mockResolvedValue({ data: { Attributes: 'mock-data' } });
modifyByOdataUrlMock.mockResolvedValue({ data: 'mock-data' });

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager extendModule isNumAttr', async () => {
  const rowMock = {
    valueType: 'int',
  };
  const res = wrapper.vm.isNumAttr(rowMock);
  expect(res).toBe(true);
});

it('test manager extendModule isFloatAttr', async () => {
  const rowMock = {
    valueType: 'float',
  };
  const res = wrapper.vm.isFloatAttr(rowMock);
  expect(res).toBe(true);
});

it('test manager extendModule isStrAttr', async () => {
  const rowMock = {
    valueType: 'string',
  };
  const res = wrapper.vm.isStrAttr(rowMock);
  expect(res).toBe(true);
});

it('test manager extendModule isBoolAttr', async () => {
  const rowMock = {
    valueType: 'bool',
  };
  const res = wrapper.vm.isBoolAttr(rowMock);
  expect(res).toBe(true);
});

it('test manager extendModule isJsonAttr', async () => {
  const rowMock = {
    valueType: 'json',
  };
  const res = wrapper.vm.isJsonAttr(rowMock);
  expect(res).toBe(true);
});

it('test manager extendModule isEditableAttr', async () => {
  const rowMock = {
    valueType: 'json',
    accessMode: 'Write',
  };
  const res = wrapper.vm.isJsonAttr(rowMock);
  expect(res).toBe(true);
});


it('test manager extendModule clickCancel', async () => {
  const odataMock = 'mock odata';
  const attrNameMock = 'mock attr'
  wrapper.vm.clickCancel(odataMock, attrNameMock);
  expect(queryByOdataUrlMock).toHaveBeenCalled();
});