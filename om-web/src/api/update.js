/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $post } from '@/api/http';

export function queryUpdateStatus(isShowLoading = true) {
  // 功能描述：查询固件升级状态信息
  const url = '/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    });
}

export function updateFirmware(params, isShowLoading = true) {
  // 功能描述：升级固件
  const url = '/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate'
  return $post(url,
    { ...params },
    {
      customParams: { isShowLoading },
    });
}

export function resetFirmware(params) {
  // 功能描述：生效固件
  const url = '/redfish/v1/UpdateService/Actions/UpdateService.Reset'
  return $post(url, {
    ...params,
  });
}

export function updateHdd(params) {
  // 功能描述：升级硬盘固件
  const url = '/redfish/v1/UpdateHddService/Actions/UpdateHddService.SimpleUpdate'
  return $post(url, {
    ...params,
  });
}

export function queryHddUpgradeInfo(hddId) {
  const url = '/redfish/v1/UpdateHddService/Actions/UpdateHddService.infos'
  let params = {
    'HddNo': parseInt(hddId),
  }
  return $post(url, { ...params });
}

export function queryUpgradeFlag() {
  const url = '/redfish/v1/UpdateHddService/upgradeFlag'
  return $get(url);
}