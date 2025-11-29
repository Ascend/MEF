/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $post } from '@/api/http';

export function queryHttpsCertInfo() {
  // 功能描述：查询SSL证书资源信息
  const url = '/redfish/v1/Systems/SecurityService/HttpsCert'
  return $get(url)
}

export function importServerCertificate(params) {
  // 功能描述：导入服务器证书
  const url = '/redfish/v1/Systems/SecurityService/HttpsCert/Actions/HttpsCert.ImportServerCertificate'
  return $post(url, {
    ...params,
  });
}

export function downloadCSRFile(params) {
  // 功能描述：下载csr
  const url = '/redfish/v1/Systems/SecurityService/downloadCSRFile'
  return $post(url, params);
}