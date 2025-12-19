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

export default (err) => {
  const { response } = err;
  if (!response.status) {
    err.code = '';
    err.message = 'no response.status';
  }
  err.code = response.status;
  err.status = response.data.status;
  switch (response.status) {
    case 200:
      err.message = '错误响应';
      break;
    case 400:
      err.message = '请求错误(400)';
      break;
    case 401:
      err.message = '未授权,请重新登陆(401)';
      break;
    case 403:
      err.message = '拒绝访问(403)';
      break;
    case 404:
      err.message = '请求出错(404)';
      break;
    case 408:
      err.message = '请求超时(408)';
      break;
    case 500:
      err.message = '服务器错误(500)';
      break;
    case 501:
      err.message = '服务器未实现(501)';
      break;
    case 502:
      err.message = '网络错误(502)';
      break;
    case 503:
      err.message = '服务不可用(503)';
      break;
    case 504:
      err.message = '网络超时(504)';
      break;
    case 505:
      err.message = 'HTTP版本不受支持(505)';
      break;
    default:
      err.message = `连接出错,状态码：(${err.response.status})!`;
  }
  return err;
};
