#!/bin/bash

# Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
# Description: 构建后去除DT使用的开源软件依赖脚本
set -e
BUILD_DEPENDENCY_CACHE_DIR="/tmp/.next3rd/"
GO_MOD_GRAPH_FILE_NAME="go-mod.graph"

# 目录不存在说明是MR流水线，跳过替换
if ! [ -d "$BUILD_DEPENDENCY_CACHE_DIR" ]; then
    exit 0
fi
find "$BUILD_DEPENDENCY_CACHE_DIR" -type f -name "$GO_MOD_GRAPH_FILE_NAME" \
    -exec chmod +w {} \; -exec sed -i '/gomonkey/d' {} \; -exec sed -i '/goconvey/d' {} \; -exec sed -i '/assertions/d' {} \;
