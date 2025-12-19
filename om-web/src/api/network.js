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

import {$get, $patch} from '@/api/http';

export function queryAllEthernetInfo(isShowLoading = true) {
  // 功能描述：查询以太网接口集合信息
  const url = '/redfish/v1/Systems/EthernetInterfaces'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    })
}

export function querySingleEthernetInfo(params) {
  // 功能描述：查询以太网接口资源信息
  const url = '/redfish/v1/Systems/EthernetInterfaces/' + params
  return $get(url);
}

export function queryWirelessStatusInfo(isShowLoading = true) {
  // 功能描述：查询无线网络接口状态资源信息
  const url = '/redfish/v1/Systems/LTE/StatusInfo'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    })
}

export function configWirelessStatusInfo(params) {
  // 功能描述：配置无线网络接口状态资源信息
  const url = '/redfish/v1/Systems/LTE/StatusInfo'
  return $patch(url,
    { ...params }
  )
}

export function queryWirelessConfigInfo() {
  // 功能描述：查询无线网络 APN接口资源信息
  const url = '/redfish/v1/Systems/LTE/ConfigInfo'
  return $get(url)
}

export function configWireless(params) {
  // 功能描述：配置无线网络 APN接口资源信息
  const url = '/redfish/v1/Systems/LTE/ConfigInfo'
  return $patch(url,
    { ...params }
  )
}
