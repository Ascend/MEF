#!/bin/bash
# Copyright © Huawei Technologies Co., Ltd. 2025. All rights reserved.

# 程序文件 & 程序文件目录
function chmod_550() {
  find "$OUTPUT_DIR" \( -name "*.so*" -o -name "*.py" -o -name "*.pyi" -o -name "*.css" \) -exec chmod 550 {} +
  find "$OUTPUT_DIR" \( -name "*.html" -o -name "*.js" -o -name "*.sh" \) -exec chmod 550 {} +

  find "$OUTPUT_DIR" \( -name "script" -o -name "lib" -o -name "bin" \) -type d -exec chmod 550 {} +
}

# 密钥组件 & 私钥 & 证书 & 加密密文
function chmod_600() {
  find "$OUTPUT_DIR" \( -name "*.cms" -o -name "*.crl" \) -exec chmod 600 {} +
}

# 配置文件 & Debug文件 & 业务数据文件
function chmod_640() {
  find "$OUTPUT_DIR" \( -name "*.xml" -o -name "*.json" -o -name "*.ttl" -o -name "*.conf" \) -exec chmod 640 {} +
  find "$OUTPUT_DIR" \( -name "*.png" -o -name "*.typed" -o -name "*.types" -o -name "*.svg" \) -exec chmod 640 {} +
  find "$OUTPUT_DIR" \( -name "*.default" -o -name "*.ini" -o -name "*.ico" -o -name "*.md" \) -exec chmod 640 {} +
  find "$OUTPUT_DIR" \( -name "*.gif"  \) -exec chmod 640 {} +

  find "$OUTPUT_DIR" \( -name "*.dat" -o -name "*.gz" -o -name "*.txt" \) -exec chmod 640 {} +
  find "$OUTPUT_DIR" \( -name "*.service" -o -name "*.target" -o -name "*.tar.gz" \) -exec chmod 640 {} +
}

# 用户主目录 & 配置文件目录 & 日志文件目录 & Debug文件目录 & 临时文件目录
function chmod_750() {
  find "$OUTPUT_DIR" -type d -exec chmod 750 {} + # 默认目录权限为750
}

function main() {
  OUTPUT_DIR=$1
  if [ -z "$OUTPUT_DIR" ] || [ ! -d "$OUTPUT_DIR" ]; then
    echo "Error: OUTPUT_DIR is either empty or does not exist."
    exit 1
  fi
  chmod_750
  chmod_640
  chmod_600
  chmod_550
}

main "$@"
exit 0
