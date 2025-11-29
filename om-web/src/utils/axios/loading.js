/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

import { ElLoading } from 'element-plus';

let loadingCount = 0;
let loading;

export const showLoading = () => {
  if (loadingCount === 0) {
    loading = ElLoading.service({
      lock: true,
      background: 'rgba(0, 0, 0, 0.6)',
    });
  }
  loadingCount += 1;
};

export const hideLoading = () => {
  if (loadingCount <= 0) {
    return;
  }
  loadingCount -= 1;
  if (loadingCount === 0) {
    loading.close();
  }
};
