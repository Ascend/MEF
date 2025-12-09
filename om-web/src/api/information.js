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

export function queryAllExtendedDevicesInfo(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询外部设备集合信息
  const url = '/redfish/v1/Systems/ExtendedDevices'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function modifySystemsSourceInfo(params, isShowLoading = true) {
  // 功能描述：修改系统资源属性
  const url = '/redfish/v1/Systems'
  return $patch(url,
    { ...params },
    {
      customParams: { isShowLoading },
      timeout: 3 * 60 * 1000,
    }
  )
}

export function querySystemsSourceInfo(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询系统资源信息
  const url = '/redfish/v1/Systems'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function queryCpuInfo(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询 CPU 资源信息
  const url = '/redfish/v1/Systems/Processors/CPU'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function queryMemoryInfo(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询内存资源信息
  const url = '/redfish/v1/Systems/Memory'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function queryAiProcessorInfo(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询AI处理器资源信息
  const url = '/redfish/v1/Systems/Processors/AiProcessor'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}