/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $patch } from '@/api/http';

export function queryAlarmSourceService(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询告警资源服务
  const url = '/redfish/v1/Systems/Alarm/AlarmInfo'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function queryAlarmShieldRules() {
  // 功能描述：查询告警屏蔽规则
  const url = '/redfish/v1/Systems/Alarm/AlarmShield'
  return $get(url)
}

export function createAlarmShieldRules(params) {
  // 功能描述：创建告警屏蔽规则
  const url = '/redfish/v1/Systems/Alarm/AlarmShield/Increase'
  return $patch(url, {
    ...params,
  })
}

export function cancelAlarmShieldRules(params) {
  // 功能描述：取消告警屏蔽规则
  const url = '/redfish/v1/Systems/Alarm/AlarmShield/Decrease'
  return $patch(url, {
    ...params,
  })
}
