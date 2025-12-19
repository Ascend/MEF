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

export default (response) => {
  const status = response.status;
  if ((status >= 200 && status <= 300) || status === 304) {
    return response;
  } else {
    const code = parseInt(response.data && response.data.code);
    let message = (response.data ?? {}).msg;

    switch (code) {
      case 400:
        message = message ?? '请求错误';
        break;
      case 401:
        message = message ?? '未授权';
        break;
      case 403:
        message = message ?? '未登录';
        break;
      case 404:
        message = message ?? '请求地址错误';
        break;
      case 412:
        message = message ?? '未找到有效session';
        break;
      default:
        break;
    }
    return {
      code,
      message,
    };
  }
};
