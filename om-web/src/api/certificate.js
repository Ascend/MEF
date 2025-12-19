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