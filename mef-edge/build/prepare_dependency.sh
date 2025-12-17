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

echo "------------------------ start prepare dependency ------------------------"

CUR_DIR=$(dirname "$(readlink -f "$0")")
MEF_EDGE_DIR=$(readlink -f "$CUR_DIR"/../)
OPENSOURCE_DIR=$MEF_EDGE_DIR/opensource
OPENSSL_PATH=${OPENSOURCE_DIR}/openssl
KUBEEDGE_PATH=${OPENSOURCE_DIR}/kubeedge
GLIBC_PATH=${OPENSOURCE_DIR}/glibc

# create opensource directory
if [ ! -d "${OPENSOURCE_DIR}" ]; then
    echo "create opensource directory ..."
    mkdir "${OPENSOURCE_DIR}"
fi

# download openssl
if [ ! -d "${OPENSSL_PATH}" ]; then
    echo "=========== start to download openssl ==========="
    git clone -b openssl-3.0.9 --depth=1 https://gitcode.com/openssl/openssl.git "${OPENSSL_PATH}"
fi

# download kubeedge
if [ ! -d "${KUBEEDGE_PATH}" ]; then
    echo "=========== start to download kubeedge ==========="
    git clone -b v1.12.6 --depth=1 https://gitcode.com/kubeedge/kubeedge.git "${KUBEEDGE_PATH}"
fi

# download glibc
if [ ! -d "${GLIBC_PATH}" ]; then
    echo "=========== start to download glibc ==========="
    git clone -b openEuler-22.03-LTS-SP3 --depth=1 https://gitcode.com/src-openeuler/glibc.git "${GLIBC_PATH}"
fi

echo "=========== download opensource package done ==========="
echo "------------------------ end prepare dependency ------------------------"