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
CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$(readlink -f "${CUR_DIR}"/../)
CPPLIB_DIR="${TOP_DIR}"/platform/cpp
PLATFORM="${TOP_DIR}"/platform
OUTPUT_LIB="${TOP_DIR}"/output/lib
OUTPUT_INC="${TOP_DIR}"/output/include
OUTPUT_BIN="${TOP_DIR}"/output/bin

function init_env() {
    mkdir -p "${OUTPUT_LIB}"
    mkdir -p "${OUTPUT_INC}"
    mkdir -p "${OUTPUT_BIN}"

    if [ ! -d "${PLATFORM}"/HuaweiSecureC ]; then
        echo "init enviroment failed, HuaweiSecureC not found."
        exit 1
    else
        if [ ! -d "${PLATFORM}"/cpp/secure ]; then
            cp -rf "${PLATFORM}"/HuaweiSecureC "${PLATFORM}"/cpp/
            mv "${PLATFORM}"/cpp/HuaweiSecureC "${PLATFORM}"/cpp/secure
        fi
    fi
}

function make_c_code()
{
    cd "${CPPLIB_DIR}"/secure/src
    make clean;make -j8
    cp -rf "${CPPLIB_DIR}"/secure/include/* "${OUTPUT_INC}"
    cp -f "${CPPLIB_DIR}"/secure/lib/libsecurec.so "${OUTPUT_LIB}"
    cd - > /dev/null

    #add security_service build
    if [ ! -f "${TOP_DIR}"/platform/SecurityService/lib.tar ];then
        echo "SecurityService lib is not exist"
        exit 1
    fi
    tar -xf "${TOP_DIR}"/platform/SecurityService/lib.tar -C "${TOP_DIR}"/platform/SecurityService
    cp -rf "${TOP_DIR}"/platform/SecurityService/lib/* "${OUTPUT_LIB}" || exit $?
    cp -rf "${TOP_DIR}"/platform/SecurityService/include/* "${OUTPUT_INC}" || exit $?
    rm -rf "${TOP_DIR}"/platform/SecurityService/lib.tar

    #build with cmake
    mkdir -p "${TOP_DIR}"/src/build
    cd "${TOP_DIR}"/src/build
    cmake ..
    make
    cp "${TOP_DIR}"/src/build/cpp/cms_verify/libverify.so "${OUTPUT_LIB}"
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


function main()
{
    init_env
    make_c_code
}

main
exit 0
