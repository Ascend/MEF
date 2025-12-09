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

import axios from 'axios';
import { showLoading, hideLoading } from '@/utils/axios/loading';
import { showErrorAlert } from '@/utils/commonMethods';
import handleResponse from '@/utils/axios/response'
import handleError from '@/utils/axios/error'
import setConfig from '@/utils/axios/config'
import { handleResponseError } from '@/utils/axios/handleResponseError'
import i18n from '@/utils/locale';

export const intactRequest = setConfig(axios);
export const request = intactRequest.create();

request.interceptors.request.use(
  (config) => {
    if (config.headers && Object.prototype.hasOwnProperty.call(config.headers, 'AutoRefresh')) {
      config.headers.AutoRefresh = String(config.headers.AutoRefresh);
    }
    request.config = Object.assign({}, config);
    if (config?.customParams.isShowLoading) {
      showLoading();
    }
    return config;
  },
  (error) => {
    if (error.config.customParams.isShowLoading) {
      showLoading();
    }
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    if (response.config?.customParams.isShowLoading) {
      hideLoading();
    }
    return Promise.resolve(handleResponse(response))
  },
  (err) => {
    if (!err.config || err.config?.customParams.isShowLoading) {
      hideLoading();
    }

    if (!err) {
      return Promise.reject(err);
    }

    if (err.response) {
      let error = handleError(err)
      handleResponseError(error.response);
    } else {
      if (axios.isCancel(err)) {
        throw new axios.Cancel(
          err.message ?? `请求'${request.config.url}'被取消`
        )
      } else if (err.stack && err.stack.includes('timeout')) {
        err.message = '请求超时';
        showErrorAlert(i18n.global.t('errorTip.requestTimeout'));
      } else {
        err.message = '连接服务器失败'
        showErrorAlert(i18n.global.t('errorTip.connectionServerFailed'));
      }
    }
    return Promise.reject(err)
  }
)