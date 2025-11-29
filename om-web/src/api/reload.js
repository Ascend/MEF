/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $post } from '@/api/http';

export function resetSystem(params) {
  // 功能描述：复位系统操作
  const url = '/redfish/v1/Systems/Actions/ComputerSystem.Reset';
  return $post(url, {
    ...params,
  });
}
