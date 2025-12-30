#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MEF is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

CUR_DIR=$(dirname $(readlink -f "$0"))
GLIBC_DIR=$(readlink -f ${CUR_DIR}/../../../opensource/glibc)
BUILD_DIR=$(readlink -f ${CUR_DIR}/../../../build)
KUBERNETES_VERSION="v1.28.1"
pushd "${CUR_DIR}" || exit

mkdir -p ${CUR_DIR}/bin
go get -v k8s.io/kubernetes@"$KUBERNETES_VERSION"
PAUSE_DIR="$(go env GOMODCACHE)/k8s.io/kubernetes@${KUBERNETES_VERSION}/build/pause/linux"
rm -f go.mod go.sum

cp ${CUR_DIR}/CMakeLists.txt ${CUR_DIR}/bin
cp ${PAUSE_DIR}/pause.c ${CUR_DIR}/bin

pushd "${CUR_DIR}"/bin || exit
cmake CMakeLists.txt
make -j8
if [[ $? -ne 0 ]]; then
    echo "compile pause failed"
    exit 1
fi
popd || exit

ARCH=$(arch 2>&1)

LD_PATH=$(ldd ./bin/pause | grep 'ld-linux' | awk 'FNR==1{print $1}')

function build_glibc()
{
    echo "-----------start build glibc ------------------------"
    mkdir -p "$GLIBC_DIR"/ascend-ci/build
    if ! cp -f "${BUILD_DIR}"/build_glibc.sh "$GLIBC_DIR"/ascend-ci/build; then
        echo "failed to copy build_glibc.sh"
        exit 1
    fi
    pushd "$GLIBC_DIR"/ascend-ci/build || exit
    if ! bash "$GLIBC_DIR"/ascend-ci/build/build_glibc.sh; then
        echo "build glibc failed"
        popd || exit
        exit 1
    fi
    popd || exit
    echo "-----------end build glibc ------------------------"
}

build_glibc

mkdir lib
cp ${GLIBC_DIR}/ascend-ci/output/lib/libc.so ./lib
cp ${GLIBC_DIR}/ascend-ci/output/lib/ld.so ./lib
chmod 755 ./lib/*

PAUSE_TAR_PATH=${CUR_DIR}/pause.tar.gz

if [ -f "${PAUSE_TAR_PATH}" ]; then
    rm -f ${PAUSE_TAR_PATH}
fi

function build_pause_image() {
    docker build --build-arg ARCH=${ARCH} --build-arg LD_PATH=${LD_PATH} --build-arg BUILD_TIME="2022-12-23 11:34:30" -f Dockerfile -t k8s.gcr.io/pause:latest .
    docker save k8s.gcr.io/pause:latest | gzip > ${CUR_DIR}/pause.tar.gz
    if [[ $? -ne 0 ]]; then
        return 1
    fi

    echo "build pause success"
    return 0
}

echo "start to build pause image"

max_attempts=5
attempt=0
while [ "${attempt}" -lt "${max_attempts}" ]; do
    if build_pause_image; then
        break
    else
        echo "build pause failed, attempt $((attempt + 1)) of ${max_attempts}, retrying in 5 seconds..."
        sleep 5
        ((attempt++))
    fi
done

if [ "${attempt}" -eq "${max_attempts}" ]; then
    echo "build pause failed after ${max_attempts} attempts"
    exit 1
fi

rm -rf "${CUR_DIR:?}"/bin/*
rm -rf "${CUR_DIR:?}"/lib/*
