/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
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
