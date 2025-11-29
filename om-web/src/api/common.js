/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $patch } from '@/api/http';
import constants from '@/utils/constants';

export function queryByOdataUrl(url, isShowLoading = true, AutoRefresh = false) {
  // 功能描述：通用查询。可直接传入 odata.id
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    });
}

export function modifyByOdataUrl(url, params, timeout = constants.DEFAULT_TIMEOUT, isShowLoading = true) {
  // 功能描述：通用修改，直接调用 odata.id
  return $patch(url,
    { ...params },
    {
      customParams: { isShowLoading },
      timeout,
    }
  )
}

export async function fetchJson(jsonName, isShowLoading = false) {
  // 功能描述：获取json
  let { data } = await $get(jsonName,
    {},
    {
      customParams: { isShowLoading },
    });
  return data;
}