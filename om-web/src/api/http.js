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

import {request} from '@/utils/axios';
import {clearSessionStorage, getAuthToken} from '@/utils/commonMethods';

const handleResponseError = (err) => {
  if (err.code === '401') {
    location.reload();
    clearSessionStorage();
  }
}

export function getUrls(params) {
  const requests = params.map(item => request.get(item.url, {
    headers: {
      ...getAuthToken(),
      AutoRefresh: item?.AutoRefresh ?? false,
    },
    customParams: {
      isShowLoading: item?.isShowLoading ?? true,
    },
  }));
  return Promise.allSettled(requests);
}

export function $get(url, params = {}, config = {}) {
  const { headers, ...rest } = config;
  return new Promise((resolve, reject) => {
    request
      .get(url, {
        params,
        headers: {
          ...getAuthToken(),
          ...headers,
        },
        ...rest,
      })
      .then(res => {
        resolve(res)
      })
      .catch(err => {
        reject(err)
        handleResponseError(err);
      })
  })
}

export function $post(url, params = {}, config = {}) {
  const { headers, ...rest } = config;
  return new Promise((resolve, reject) => {
    request
      .post(url, params, {
        headers: {
          ...getAuthToken(),
          ...headers,
        },
        ...rest,
      })
      .then(res => {
        resolve(res)
      })
      .catch(err => {
        reject(err)
        if (url !== '/redfish/v1/SessionService/Sessions') {
          handleResponseError(err);
        }
      })
  })
}

export function $delete(url, params = {}, config = {}) {
  const { headers, ...rest } = config;
  return new Promise((resolve, reject) => {
    request
      .delete(url, {
        params,
        headers: {
          ...getAuthToken(),
          ...headers,
        },
        ...rest,
      })
      .then(res => {
        resolve(res)
      })
      .catch(err => {
        reject(err)
        handleResponseError(err);
      })
  })
}

export function $patch(url, params = {}, config = {}) {
  return new Promise((resolve, reject) => {
    request
      .patch(url, params, {
        headers: {
          ...getAuthToken(),
        },
        ...config,
      })
      .then((data) => {
        resolve(data);
      })
      .catch((err) => {
        reject(err);
        handleResponseError(err);
      });
  });
}

