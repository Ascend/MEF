#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MEF is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

OUTPUT_DIR=""

function chmod_400() {
  find "$OUTPUT_DIR" -name "*.xml" -print0 | xargs -0 chmod 400
  find "$OUTPUT_DIR" -name "*.so*" -print0 | xargs -0 chmod 400
}

function chmod_500() {
  find "$OUTPUT_DIR" -type f -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "bin" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "script" -type d -print0 | xargs -0 chmod 500
  find "$OUTPUT_DIR" -name "lib" -type d -print0 | xargs -0 chmod 500
}

function chmod_600() {
  find "$OUTPUT_DIR" -name "*.json" -print0 | xargs -0 chmod 600
  find "$OUTPUT_DIR" -name "*.service" -print0 | xargs -0 chmod 600
  find "$OUTPUT_DIR" -name "*.target" -print0 | xargs -0 chmod 600
  find "$OUTPUT_DIR" -name "*.tar.gz" -print0 | xargs -0 chmod 600
}

function chmod_700() {
  find "$OUTPUT_DIR" -type d -print0 | xargs -0 chmod 700
}

function main() {
  OUTPUT_DIR=$1
  if [ -z "$OUTPUT_DIR" ]; then
    return 0
  fi
  chmod_700
  chmod_500
  chmod_400
  chmod_600
}

main "$@"
