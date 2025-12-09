#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
