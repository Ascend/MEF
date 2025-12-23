#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
set -e

CUR_DIR=$(dirname $(readlink -f "$0"))
CI_DIR=$(readlink -f $CUR_DIR/../)
OPENSOURCE_DIR=${CI_DIR}/../../
TOP_DIR=${CI_DIR}/../../../

pushd $OPENSOURCE_DIR/libboundscheck

make clean
make -j8
popd

mkdir -p "${TOP_DIR}/output/include"
mkdir -p "${TOP_DIR}/output/lib"
cp -rf "${OPENSOURCE_DIR}"/libboundscheck/include/* "${TOP_DIR}/output/include"
cp -f "${OPENSOURCE_DIR}"/libboundscheck/lib/libboundscheck.so "${TOP_DIR}/output/lib/libsecurec.so"
