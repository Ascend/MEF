/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import i18n from '@/utils/locale';

/**
 * MAC地址:支持两种格式，xx:xx:xx 和 xx:xx:xx:xx:xx:xx
 */
export const macChecker = (rule, value, callback) => {
  let mac1 = new RegExp('^([A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2}$');
  let mac2 = new RegExp('^([A-Fa-f0-9]{2}:){2}[A-Fa-f0-9]{2}$');
  if (value === '' || mac1.test(value) || mac2.test(value)) {
    callback();
    return;
  }
  callback(new Error(i18n.global.t('safety.loginRules.macErrorTip')));
};

export const ipWithMaskChecker = (rule, value, callback) => {
  if (value && value.length > 18) {
    callback(new Error(i18n.global.t('safety.loginRules.ipErrorTip')));
    return;
  }

  let ip1 = new RegExp(
    /^((25[0-5])|(2[0-4]\d)|(1\d\d)|([1-9]\d)|[1-9])(\.((25[0-5])|(2[0-4]\d)|(1\d\d)|([1-9]\d)|\d)){3}\/(([12][0-9]?)|(3[0-2])|([3-9]))$/
  );
  let ip2 = new RegExp(
    /^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$/
  );
  if (value === '' || ip1.test(value) || ip2.test(value)) {
    callback();
    return;
  }
  callback(new Error(i18n.global.t('safety.loginRules.ipErrorTip')));
};

/**
 * common ip validate function.
 */
export function validateIp(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  if (value && value.length > 15) {
    callback(new Error(i18n.global.t('common.errorFormatTip')));
    return;
  }
  const pattern = /^(((\d)|([1-9]\d)|(1\d{2})|(2[0-4]\d)|(25[0-5]))\.){3}((\d)|([1-9]\d)|(1\d{2})|(2[0-4]\d)|(25[0-5]))$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('common.errorFormatTip')))
  } else {
    callback()
  }
}

/**
* validate net management configuration's node id field.
*/
export function validateNodeId(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }

  const pattern = /^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('registration.nms.nodeIdErrorTip')))
  } else {
    callback()
  }
}

/**
* validate net management configuration's port field.
*/
export function validatePort(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }

  const pattern = /^[0-9]+$/;
  if (!pattern.exec(value) || value < 1 || value > 65535) {
    callback(new Error(i18n.global.t('registration.nms.portErrorTip')))
  } else {
    callback()
  }
}

/**
* validate net management configuration's server name field.
*/
export function validateServerName(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }

  const pattern = /^[A-Za-z0-9-.]{0,64}$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('registration.nms.serverNameErrorTip')))
  } else {
    callback()
  }
}

/**
* validate net management configuration's account field.
*/
export function validateAccount(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }

  const pattern = /^[0-9a-zA-Z-_]{1,256}$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('registration.nms.accountErrorTip')))
  } else {
    callback()
  }
}

/**
* validate HostName field in system information page.
*/
export function validateHostName(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  let pattern = /^(?!-)[A-Za-z0-9-]{1,63}(?<!-)$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('information.systemInfo.hostnameErrTip')))
  } else {
    callback()
  }
}

/**
* validate AssetTag field in system information page.
*/
export function validateAssetTag(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  const pattern = /^([\x20-\x7E]){1,255}$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('information.systemInfo.AssetTagErrTip')))
  } else {
    callback()
  }
}


/**
* validate session timeout field in safety page.
*/
export function validateSessionTimeout(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  if (value < 5 || value > 120) {
    callback(new Error(i18n.global.t('common.errorRangeTip')))
  } else {
    callback()
  }
}


/**
* validate cert alarm time field in safety page.
*/
export function validateCertAlarmTime(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  if (value < 7 || value > 180) {
    callback(new Error(i18n.global.t('common.errorRangeTip')))
  } else {
    callback()
  }
}


/**
* validate password expiration days field in safety page.
*/
export function validatePasswordExpirationDays(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  if (value < 0 || value > 365) {
    callback(new Error(i18n.global.t('common.errorRangeTip')))
  } else {
    callback()
  }
}

/**
* validate usage field in the network-wired config form
*/
export function validateNetworkUsage(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  const pattern = /^[0-9a-zA-Z_]{1,32}$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('network.networkWired.ipUsageErrorTip')))
  } else {
    callback()
  }
}


/**
* common subnet mask validator
*/
export function validateVlanId(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  const pattern = /^\d+$/;
  if (!pattern.test(value) || value < 1 || value > 4094) {
    callback(new Error(i18n.global.t('network.networkWired.vlanIdErrorTip')))
  } else {
    callback()
  }
}

/**
* validate ServerDir & MountPath field in nfs page.
*/
export function validateServerPath(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  const pattern = /^\/[0-9a-zA-Z/_-]{1,255}$/;
  if (!pattern.test(value)) {
    callback(new Error(i18n.global.t('disk.common.serverPathTip')))
  } else {
    callback()
  }
}

/**
* validate ApnName field in Wireless page.
*/
export function validateApnName(rule, value, callback) {
  if (!value) {
    callback()
    return;
  }
  let pattern = new RegExp(/^[a-zA-Z0-9\-_\.@]{1,39}$/);
  if (!value.match(pattern)) {
    callback(new Error(i18n.global.t('network.networkWireless.apnNameErrorTip')))
  } else {
    callback()
  }
}

/**
+* validate float field in extend modules page.
+*/
export function validateFloat(value) {
  let pattern = /^(-?\d+)(\.\d+)?$/;
  return pattern.test(value) ? null : i18n.global.t('extendModule.floatErrorTip')
}

/**
+* validate non-float number in extend modules page.
+*/
export function validateNonFloatNumber(value) {
  let pattern = /^-?\d+$/;
  return pattern.test(value) ? null : i18n.global.t('extendModule.intErrorTip')
}
