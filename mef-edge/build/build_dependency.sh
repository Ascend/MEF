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

set -e

echo "------------------------ start build dependency ------------------------"

CUR_DIR=$(dirname "$(readlink -f "$0")")
MEF_EDGE_DIR=$(readlink -f "$CUR_DIR"/../)

rm -rf "${MEF_EDGE_DIR}/output"
if ! mkdir -p "${MEF_EDGE_DIR}/output"; then
    echo "create output dir ${MEF_EDGE_DIR}/output failed"
    exit 1
fi

BUILD_SCRIPT="${MEF_EDGE_DIR}/build/build_c_package.sh"

if [ ! -f "${BUILD_SCRIPT}" ]; then
    echo "build script ${BUILD_SCRIPT} not found"
    exit 1
fi

chmod +x "${BUILD_SCRIPT}"

if ! bash "${BUILD_SCRIPT}"; then
    echo "execute script ${BUILD_SCRIPT} failed"
    exit 1
fi
echo "execute script ${BUILD_SCRIPT} success"
echo "------------------------ end build dependency ------------------------"