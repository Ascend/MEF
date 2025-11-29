#!/bin/bash
# Copyright © Huawei Technologies Co., Ltd. 2020-2025. All rights reserved.
CUR_DIR=$(dirname $(readlink -f "$0"))
CI_DIR=$(readlink -f $CUR_DIR/../)
OPENSOURCE_DIR=${CI_DIR}/../../
ROOT_DIR=$(readlink -f $CI_DIR/../)

if [ ! -d "${OPENSOURCE_DIR}"/openssl ];then
    echo "openssl package is not exist!"
    exit 1
fi

if [ -d "${CI_DIR}"/output ];then
    echo "openssl is already compiled."
    exit 0
fi

cd ${ROOT_DIR}


${ROOT_DIR}/config shared --prefix=/ CFLAGS="${CFLAGS_ENV}" CXXFLAGS="${CXXFLAGS}"

make -j8

if [[ $? -ne 0 ]];then
    echo "build openssl failed!"
    exit 1
fi

mkdir -p ${CI_DIR}/output/
mkdir -p ${CI_DIR}/output/bin/

cp ${ROOT_DIR}/libcrypto.so* ${CI_DIR}/output -d
cp ${ROOT_DIR}/libssl.so* ${CI_DIR}/output -d
cp -rf ${ROOT_DIR}/include ${CI_DIR}/output

mkdir -p "${CI_DIR}"/output/include/crypto/rsa/
cp -f "${ROOT_DIR}"/crypto/rsa/rsa_local.h "${CI_DIR}"/output/include/crypto/rsa/

mkdir -p "${CI_DIR}"/output/include/crypto/evp/
cp -f "${ROOT_DIR}"/crypto/evp/evp_local.h "${CI_DIR}"/output/include/crypto/evp/


cp apps/openssl ${CI_DIR}/output/bin
cp apps/openssl.cnf ${CI_DIR}/output

echo "build openssl successfully!"

exit 0
