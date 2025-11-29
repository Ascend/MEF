/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $post } from '@/api/http';

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
