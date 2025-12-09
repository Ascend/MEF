/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   MindEdge is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

import { $get, $post } from '@/api/http';

export function queryEthernetIpList() {
  // 功能描述：获取网口与IP列表
  const url = '/redfish/v1/Systems/EthIpList'
  return $get(url)
}

export function resetAction(params) {
  // 功能描述：恢复出厂设置
  const url = '/redfish/v1/Systems/Actions/RestoreDefaults.Reset'
  return $post(url, {
    ...params,
  })
}

export function restoreDefaults(params) {
  // 功能描述：恢复默认配置
  const url = '/redfish/v1/Systems/Actions/RestoreDefaults.Config'
  return $post(url, {
    ...params,
  })
}