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
# Description: Mef  相关c库打包脚本

set -e

CUR_DIR=$(dirname "$(readlink -f "$0")")
MEF_EDGE_DIR=$(readlink -f "$CUR_DIR"/../)
OPENSOURCE_DIR=${MEF_EDGE_DIR}/opensource

function prepare() {
    cd "$MEF_EDGE_DIR"
    if [ ! -d "output/lib" ];then
        mkdir -p "output/lib"
    else
        rm -rf "output/lib/*"
    fi

    if [ ! -d "output/include" ]; then
        mkdir -p "output/include"
    else
        rm -rf output/include/*
    fi
}

function build_opensource()
{
    declare opensources=(openssl)
    for taskname in "${opensources[@]}"
    do
        echo "-----------start build ${taskname} ------------------------"
        mkdir -p "$OPENSOURCE_DIR"/"${taskname}"/ascend-ci/build
        cp -f "${CUR_DIR}"/build_"${taskname}".sh "$OPENSOURCE_DIR"/"${taskname}"/ascend-ci/build
        cd "$OPENSOURCE_DIR"/"${taskname}"/ascend-ci/build
        bash "$OPENSOURCE_DIR"/"${taskname}"/ascend-ci/build/build_"${taskname}".sh
        if [[ $? != 0 ]];then
            echo build ${taskname} failed
            exit 1
        fi
        echo "-----------end build ${taskname} ------------------------"
    done

    cp -f -d "$OPENSOURCE_DIR"/openssl/ascend-ci/output/libcrypto.so* "$MEF_EDGE_DIR"/output/lib
    cp -f -d "$OPENSOURCE_DIR"/openssl/ascend-ci/output/libssl.so* "$MEF_EDGE_DIR"/output/lib
    cp "$OPENSOURCE_DIR"/openssl/ascend-ci/output/include/*  "$MEF_EDGE_DIR"/output/include -rf
}

function main()
{
    prepare
    build_opensource
}

main
