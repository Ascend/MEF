/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2023. All rights reserved.
 */

export const saveFile = (data, fileName) => {
  const blob = new Blob([data], { type: 'application/octet-stream' });
  const a = document.createElement('a');
  const url = window.URL.createObjectURL(blob);

  a.href = url;
  a.download = fileName;
  a.style.display = 'none';
  document.body.appendChild(a);
  a.click();
  a.parentNode.removeChild(a);
  window.URL.revokeObjectURL(url);
}
