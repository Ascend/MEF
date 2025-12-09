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
import constants from '@/utils/constants';

export function queryLogsInfo() {
  // 功能描述：查询日志服务集合资源信息
  const url = '/redfish/v1/Systems/LogServices'
  return $get(url)
}

export function downloadLogInfo(params) {
  // 功能描述：下载日志信息
  const url = '/redfish/v1/Systems/LogServices/Actions/download'
  return $post(url,
    { ...params },
    {
      timeout: 5 * constants.MINUTE_TIMEOUT,
      responseType: 'blob',
    });
}

export function queryLogCollectProgress() {
  // 功能描述：查询日志服务集合资源信息
  const url = '/redfish/v1/Systems/LogServices/progress'
  return $get(url)
}
