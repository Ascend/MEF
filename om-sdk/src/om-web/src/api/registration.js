/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
   OMSDK is licensed under Mulan PSL v2.
   You can use this software according to the terms and conditions of the Mulan PSL v2.
   You may obtain a copy of Mulan PSL v2 at:
            http://license.coscl.org.cn/MulanPSL2
   THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
   EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
   MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
   See the Mulan PSL v2 for more details.
 */

import {$get, $post} from '@/api/http';

export function queryNetManagerInfo() {
  // 功能描述：查询网管资源信息
  const url = '/redfish/v1/NetManager';
  return $get(url);
}

export function queryNetManagerNodeID() {
  // 功能描述：查询网管节点id
  const url = '/redfish/v1/NetManager/NodeID';
  return $get(url);
}

export function queryFdRootCert() {
  // 功能描述：查询fd根证书信息
  const url = '/redfish/v1/NetManager/QueryFdCert';
  return $get(url);
}

export function modifyNetManagerInfo(params) {
  // 功能描述：配置网管资源信息
  const url = '/redfish/v1/NetManager';
  return $post(url, { ...params }, { timeout: 4.25 * 60 * 1000 });
}
