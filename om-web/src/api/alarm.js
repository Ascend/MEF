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
