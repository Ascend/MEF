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

CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$(readlink -f "${CUR_DIR}"/../)

OUTPUT_LIB="${TOP_DIR}"/output/lib

function main()
{
    #build with cmake
    mkdir -p "${TOP_DIR}"/src/build
    cd "${TOP_DIR}"/src/build
    cmake ..
    make
    cp "${TOP_DIR}"/src/build/cpp/common/libcommon.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/certmanage/libcertmanage.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/ens/ensd "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/ens/base/libbase.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/devm/libdevm.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/lpeblock/liblpeblock.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/fault_check/libfault_check.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/alarm_process/libalarm_process.so "${OUTPUT_LIB}"
    cp "${TOP_DIR}"/src/build/cpp/extend_alarm/libextend_alarm.so "${OUTPUT_LIB}"
}

main
exit 0
