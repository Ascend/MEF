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