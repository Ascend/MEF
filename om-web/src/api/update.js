/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   OMSDK is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

import {$get, $post} from '@/api/http';

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

export function queryHddInfo(hddId) {
  // 功能描述：查询硬盘固件
  const url = '/redfish/v1/UpdateHddService/HddInfo/' + hddId
  return $get(url);
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