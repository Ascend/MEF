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

export function querySystemTime(isShowLoading = true) {
  // 功能描述：查询系统时间
  const url = '/redfish/v1/Systems/SystemTime'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    })
}

export function queryNTPService(isShowLoading = true) {
  // 功能描述：查询NTP服务信息
  const url = '/redfish/v1/Systems/NTPService'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
    })
}

export function configNTPService(params, isShowLoading = true) {
  // 功能描述：设置NTP服务信息
  const url = '/redfish/v1/Systems/NTPService'
  return $patch(url,
    { ...params },
    {
      customParams: { isShowLoading },
    })
}