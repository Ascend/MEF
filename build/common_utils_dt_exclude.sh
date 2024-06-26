#!/bin/bash

# Copyright (c) 2024. Huawei Technologies Co., Ltd. All rights reserved.
# Description: 构建后去除DT使用的开源软件依赖脚本
BUILD_DEPENDENCY_CACHE_DIR="/tmp"
GO_MOD_GRAPH_FILE_NAME="go-mod.graph"

find $BUILD_DEPENDENCY_CACHE_DIR -type f -name $GO_MOD_GRAPH_FILE_NAME \
    -exec chmod +w {} \; -exec sed -i '/gomonkey/d' {} \; -exec sed -i '/goconvey/d' {} \; -exec sed -i '/assertions/d' {} \;
