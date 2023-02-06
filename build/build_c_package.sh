#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2019-2025. All rights reserved.
# Description: AtlasEdge 相关c库打包脚本

CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$(readlink -f "$CUR_DIR"/../)
PLATFORM_DIR=$TOP_DIR/platform
OPENSOURCE_DIR=$TOP_DIR/opensource
CMS_VERIFY_DIR=$TOP_DIR/MEF_Utils/cmsverifytool

function process()
{
    cd "$PLATFORM_DIR"/HuaweiSecureC/src
    make clean;make -j8
    if [ $? != 0 ];then
        echo "build secure failed"
        return 1
    fi
    cp -rf "$PLATFORM_DIR"/HuaweiSecureC/include/* "$TOP_DIR"/output/include
    cp -rf "$PLATFORM_DIR"/HuaweiSecureC/lib/libsecurec.so "$TOP_DIR"/output/lib

    cd "$CMS_VERIFY_DIR"/build
    cmake CMakeLists.txt;make -j8
    if [ $? != 0 ];then
        echo "build cms_verify failed"
        return 1
    fi
    cp "$CMS_VERIFY_DIR"/build/libcms_verify.so "$TOP_DIR"/output/lib
    cp "$CMS_VERIFY_DIR"/include/* "$TOP_DIR"/output/include
    return 0
}

function prepare() {
    cd "$TOP_DIR"
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
    cp "$OPENSOURCE_DIR"/openssl/ascend-ci/output/libcrypto.so* "$TOP_DIR"/output/lib
    cp "$OPENSOURCE_DIR"/openssl/ascend-ci/output/libssl.so* "$TOP_DIR"/output/lib
    cp "$OPENSOURCE_DIR"/openssl/ascend-ci/output/include/*  "$TOP_DIR"/output/include -rf
    return 0
}

function main()
{
    prepare
    build_opensource
    process
}

main
