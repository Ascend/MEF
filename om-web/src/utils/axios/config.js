/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import constants from '@/utils/constants';
import { getAuthToken } from '@/utils/commonMethods';

export default (axios, config = {}) => {
  const defaultConfig = {
    baseURL: '',
    timeout: constants.DEFAULT_TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
      'withCredentials': true,
      'X-Auth-Token': sessionStorage.getItem('token') ?? '',
      'AutoRefresh': false,
    },
    customParams: {
      isShowLoading: true,
    },
  }

  Object.assign(axios.defaults, defaultConfig, config);
  return axios;
}