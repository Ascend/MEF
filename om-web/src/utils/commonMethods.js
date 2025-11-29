/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { ElMessage, ElMessageBox } from 'element-plus';
import i18n from '@/utils/locale';
import errorTipMapper from '@/api/errorTipMapper';
import { fetchJson } from '@/api/common';

export function convertToGB(value) {
  return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

export function dateFormat(date, format) {
  let month = date.getMonth() + 1
  let monthStr = month.toString().padStart(2, '0')
  let result = format.replace('yyyy', date.getFullYear())
    .replace('MM', monthStr)
    .replace('dd', date.getDate().toString().padStart(2, '0'))
    .replace('hh', date.getHours().toString().padStart(2, '0'))
    .replace('mm', date.getMinutes().toString().padStart(2, '0'))
    .replace('ss', date.getSeconds().toString().padStart(2, '0'))
  return result
}

export function getAuthToken() {
  return { 'X-Auth-Token': sessionStorage.getItem('token') ?? '' };
}

export function getLanguage() {
  return sessionStorage.getItem('locale');
}

export function getLanguageDefaultChinese() {
  return sessionStorage.getItem('locale') ?? 'zh';
}

/**
* Encapsulate a customize ElMessage method with 'showClose' and 'offset' attr.
*/
let messageBasicConfig = {
  showClose: true,
  offset: 60,
  duration: 5000,
}
export const AppMessage = {
  info: (message) => {
    ElMessage({
      message,
      ...messageBasicConfig,
    })
  },
  success: (message) => {
    ElMessage.success({
      message,
      ...messageBasicConfig,
    })
  },
  error: (message) => {
    ElMessage.error({
      message,
      ...messageBasicConfig,
    })
  },
  warning: (message) => {
    ElMessage.warning({
      message,
      ...messageBasicConfig,
    })
  },
}

/**
* After request or operation error, should show alert to display error message.
*/
export function showErrorAlert(message) {
  ElMessageBox.alert(
    message,
    i18n.global.t('common.errorReminder'),
    { type: 'error' }
  )
}

/**
* show warning alert.
*/
export function showWarningAlert(message) {
  ElMessageBox.alert(
    message,
    i18n.global.t('common.tips'),
    { type: 'warning' }
  )
}

export const deepCopyDictList = (src) => {
  let dst = [];
  for (let i = 0; i < src?.length; i++) {
    dst.push({ ...src[i] });
  }
  return dst;
};

export function clearSessionAndJumpToLogin(router) {
  clearSessionStorage();
  router.push('/login');
}

export function switchToTargetTimezone(currDate, currTz, targetTz) {
  let hereNow = new Date();
  let currTimezoneDate = new Date(hereNow.toLocaleString('en-US', { timeZone: currTz }));
  let targetTimezoneDate = new Date(hereNow.toLocaleString('en-US', { timeZone: targetTz }));
  let diff = currTimezoneDate.getTime() - targetTimezoneDate.getTime();

  return new Date(currDate.getTime() - diff);
}

export function handleUploadError(error) {
  let response = JSON.parse(error?.message);
  if (!response?.error || !Object.prototype.hasOwnProperty.call(response?.error, '@Message.ExtendedInfo')) {
    showErrorAlert(i18n.global.t('common.uploadErrorTip'))
    return;
  }
  let status = response?.error['@Message.ExtendedInfo'][0]?.Oem?.status;
  if (Object.prototype.hasOwnProperty.call(errorTipMapper, status)) {
    showErrorAlert(i18n.global.t(errorTipMapper[status]))
  } else {
    showErrorAlert(i18n.global.t('common.uploadErrorTip'))
  }
}

export function isValidUploadFilename(filename) {
  let pattern = /^([a-zA-Z0-9_.-]){1,255}$/;
  let notHaveDoubleDot = filename.indexOf('..') === -1;
  return pattern.exec(filename) && notHaveDoubleDot
}

export function checkModuleType(module, type) {
  return module && module.toLowerCase().indexOf(type) !== -1
}

export function handleOperationResponseError(err, defaultTip = i18n.global.t('common.operationErrorTip')) {
  let status = err?.response?.data?.error['@Message.ExtendedInfo'][0]?.Oem?.status;
  if (Object.prototype.hasOwnProperty.call(errorTipMapper, status)) {
    showErrorAlert(i18n.global.t(errorTipMapper[status]))
  } else {
    showErrorAlert(defaultTip)
  }
}

export function clearSessionStorage() {
  let locale = sessionStorage.getItem('locale')
  sessionStorage.clear();
  sessionStorage.setItem('locale', locale)
}

export function isA500() {
  return sessionStorage.getItem('model') && sessionStorage.getItem('model')?.indexOf('500') !== -1;
}

export function isA200() {
  return sessionStorage.getItem('model') && sessionStorage.getItem('model')?.indexOf('200') !== -1;
}

export async function generateA200Routes(router) {
  if (!isA200()) {
    return;
  }

  let routesConfig = await fetchJson('/config/routesConfig.json');
  Object.keys(routesConfig).forEach(parentRouteName => {
    const parentRoute = router.options.routes.find(el => el.name === parentRouteName);
    Object.keys(routesConfig[parentRouteName]).forEach((route, index) => {
      let hasExtendModule = route === 'extendModule' && !JSON.parse(sessionStorage.getItem('hasExtendModule'))
      if (routesConfig[parentRouteName][route] && !hasExtendModule) {
        return
      }
      const routeIndex = parentRoute.children.findIndex(item => item.path === route)
      if (routeIndex !== -1) {
        router.removeRoute(`${route}.menuName`);
        parentRoute.children.splice(routeIndex, 1);
      }
    })
  })
}

export function isFulfilled(item) {
  // 判断 promise 对象是否为 fulfilled
  return item.status === 'fulfilled'
}

export function checkUrlsResponse(allResponse, AutoRefresh = false) {
  // 用于检测调用了 getUrls() 方法的接口是否有失败响应
  // 只要有一个接口请求失败就会出现错误提示
  if (!allResponse.every(item => isFulfilled(item)) && !AutoRefresh) {
    showErrorAlert(i18n.global.t('common.requestError'))
  }
}