/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { $get, $post, $patch, $delete } from '@/api/http';

export function createSession(params) {
  // 功能描述：创建会话
  const url = '/redfish/v1/SessionService/Sessions'
  return $post(url, {
    ...params,
  })
}

export function deleteSession(sessionId) {
  // 功能描述：删除会话
  const url = `/redfish/v1/SessionService/Sessions/${sessionId}`
  return $delete(url)
}

export function getUserInfo(userId) {
  // 功能描述：获取指定用户信息
  const url = `/redfish/v1/AccountService/Accounts/${userId}`;
  return $get(url)
}

export function modifyUserInfo(userId, params) {
  // 功能描述：修改指定用户信息
  const url = `/redfish/v1/AccountService/Accounts/${userId}`;
  return $patch(url, {
    ...params,
  })
}

export function querySessionService(isShowLoading = true, AutoRefresh = false) {
  // 功能描述：查询会话服务信息
  const url = '/redfish/v1/SessionService'
  return $get(url,
    {},
    {
      customParams: { isShowLoading },
      headers: { AutoRefresh },
    })
}

export function modifySessionService(params) {
  // 功能描述：修改会话服务信息
  const url = '/redfish/v1/SessionService'
  return $patch(url, {
    ...params,
  })
}

export function queryAccountService() {
  // 功能描述：查询用户服务信息
  const url = '/redfish/v1/AccountService'
  return $get(url)
}

export function modifyAccountService(params) {
  // 功能描述：修改用户服务信息
  const url = '/redfish/v1/AccountService'
  return $patch(url, {
    ...params,
  })
}

