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

import {$get, $patch, $post} from '@/api/http';
import constants from '@/utils/constants';

const SECURITYLOAD_URL = '/redfish/v1/Systems/SecurityService/SecurityLoad';

export function queryLoginRules() {
  // 功能描述：查询登录规则信息
  return $get(SECURITYLOAD_URL)
}

export function modifyLoginRules(params) {
  // 功能描述：配置登录规则信息
  return $patch(SECURITYLOAD_URL, {
    ...params,
  })
}

export function importLoginRules(params) {
  // 功能描述：导入登录规则信息
  const url = `${SECURITYLOAD_URL}/Actions/SecurityLoad.Import`
  return $post(url, {
    ...params,
  });
}

export function exportLoginRules(params) {
  // 功能描述：导出登录规则信息
  const url = `${SECURITYLOAD_URL}/Actions/SecurityLoad.Export`;
  return $post(
    url,
    { ...params },
    {
      timeout: 1 * constants.MINUTE_TIMEOUT,
    }
  );
}

export function importPunyDict(params) {
  const url = '/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictImport'
  return $post(url, params);
}

export function exportPunyDict(params) {
  const url = '/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictExport'
  return $post(url, params);
}

export function deletePunyDict(params) {
  const url = '/redfish/v1/Systems/SecurityService/Actions/SecurityService.PunyDictDelete'
  return $post(url, params);
}

export function queryHttpsCertAlarmTime() {
  // 功能描述：查询证书有效期提醒时间
  const url = '/redfish/v1/Systems/SecurityService/HttpsCertAlarmTime'
  return $get(url)
}

export function modifyHttpsCertAlarmTime(params) {
  // 功能描述：修改证书有效期提醒时间
  const url = '/redfish/v1/Systems/SecurityService/HttpsCertAlarmTime'
  return $patch(url, {
    ...params,
  })
}