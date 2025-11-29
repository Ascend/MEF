/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import 'element-plus/theme-chalk/dark/css-vars.css';
import '@/assets/css/main.css';
import i18n from '@/utils/locale';

const app = createApp(App)

app.use(router);
app.use(ElementPlus);
app.use(i18n);
app.mount('#app');

