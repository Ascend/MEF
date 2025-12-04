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
# Description: openssl 构建脚本

set -e
CUR_DIR=$(dirname $(readlink -f "$0"))
CI_DIR=$(readlink -f "$CUR_DIR"/../)
ROOT_DIR=$(readlink -f "$CI_DIR"/../)

CFLAGS_ENV="-Wall -fstack-protector-strong -fPIC -D_FORTIFY_SOURCE=2 -O2 -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack -s"
CXXFLAGS_ENV="-Wall -fstack-protector-strong -fPIC -D_FORTIFY_SOURCE=2 -O2 -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack -s"

if [ -d "${CI_DIR}"/output ];then
    echo "openssl is already compiled."
    exit 0
fi


cd "${ROOT_DIR}"
./config --prefix=/ CFLAGS="${CFLAGS_ENV}" CXXFLAGS="${CXXFLAGS_ENV}"

make -j8

if [[ $? -ne 0 ]];then
    echo "build openssl failed!"
    exit 1
fi


mkdir -p "${CI_DIR}"/output/

cp "${ROOT_DIR}"/libcrypto.so* "${CI_DIR}"/output -d
cp "${ROOT_DIR}"/libssl.so* "${CI_DIR}"/output -d
cp -rf "${ROOT_DIR}"/include "${CI_DIR}"/output

mkdir -p "${CI_DIR}"/output/include/crypto/rsa/
cp -rf "${ROOT_DIR}"/crypto/rsa/rsa_local.h "${CI_DIR}"/output/include/crypto/rsa/

mkdir -p "${CI_DIR}"/output/include/crypto/evp/
cp -rf "${ROOT_DIR}"/crypto/evp/evp_local.h "${CI_DIR}"/output/include/crypto/evp/

exit 0