/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { createI18n } from 'vue-i18n';

import zh from '@/utils/locale/zh';
import en from '@/utils/locale/en';

const i18n = createI18n({
  legacy: false, // 如果要支持compositionAPI，此项必须设置为false;
  globalInjection: true, // 全局注册$t方法
  locale: localStorage.getItem('locale') ?? 'zh',
  messages: {
    zh,
    en,
  },
});

export default i18n;
