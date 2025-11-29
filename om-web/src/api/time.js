/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $patch } from '@/api/http';

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