import { expect, it } from 'vitest';
import { shallowMount } from '@vue/test-utils';
import ElementPlus from 'element-plus';
import i18n from '@/utils/locale';

import Index from '@/views/manager/disk/Index.vue';

const wrapper = shallowMount(Index, {
  global: {
    plugins: [ElementPlus, i18n],
  },
});

it('test manager disk Index.vue', async () => {
  const activeName = wrapper.vm.activeName;
  expect(wrapper.html()).toContain(activeName);
  expect(wrapper.findComponent({ name: 'el-tabs' }).exists()).toBe(true);
});