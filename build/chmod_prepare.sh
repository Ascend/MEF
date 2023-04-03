#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
# Description: Mef  相关文件权限脚本
set -e
function chmod_400() {
  local output_dir="$1"
  find "${output_dir}" -name "*.xml" -print0 | xargs -0 chmod 400
  find "${output_dir}" -name "*.so*" -print0 | xargs -0 chmod 400
  find "${output_dir}" -name "*.html" -print0 | xargs -0 chmod 400
  find "${output_dir}" -name "*.conf" -print0 | xargs -0 chmod 400
  find "${output_dir}" -name "*.types" -print0 | xargs -0 chmod 400
  find "${output_dir}" -name "Dockerfile" -print0 | xargs -0 chmod 400
}

function chmod_500() {
  local output_dir="$1"
  find "${output_dir}" -type f -print0 | xargs -0 chmod 500
  find "${output_dir}" -name "bin" -type d -print0 | xargs -0 chmod 500
  find "${output_dir}" -name "scripts" -type d -print0 | xargs -0 chmod 500
  find "${output_dir}" -name "lib" -type d -print0 | xargs -0 chmod 500
  find "${output_dir}" -name "html" -type d -print0 | xargs -0 chmod 500
  find "${output_dir}" -name "conf" -type d -print0 | xargs -0 chmod 500
  find "${output_dir}" -name "nginx" -type d -print0 | xargs -0 chmod 500
}

function chmod_600() {
  local output_dir="$1"
  find "${output_dir}" -name "*.yaml" -print0 | xargs -0 chmod 600
  find "${output_dir}" -name "*.json" -print0 | xargs -0 chmod 600
}

function chmod_700() {
  local output_dir="$1"
  find "${output_dir}" -type d -print0 | xargs -0 chmod 700
}

function main() {
  local output_dir="$1"
  if [ -z "${output_dir}" ]; then
    return 0
  fi
  chmod_700 "${output_dir}"
  chmod_500 "${output_dir}"
  chmod_400 "${output_dir}"
  chmod_600 "${output_dir}"
}

main "$@"
