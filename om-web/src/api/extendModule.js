/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get } from '@/api/http';
  	 
export function queryAllModules(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询扩展模组集合信息
  const url = '/redfish/v1/Systems/Modules'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}