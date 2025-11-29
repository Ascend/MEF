#!/bin/bash

# Copyright (c) Huawei Technologies Co., Ltd. 2023-2025. All rights reserved.
# Description: glibc 构建脚本

set -e
CUR_DIR=$(dirname $(readlink -f "$0"))
CI_DIR=$(readlink -f "$CUR_DIR"/../)
ROOT_DIR=$(readlink -f "$CI_DIR"/../)

CFLAGS_ENV="-Wall -fstack-protector-strong -fPIC -D_FORTIFY_SOURCE=2 -O2 -Wl,-z,relro -Wl,-z,now -Wl,-z,noexecstack -s"

rm -rf "${ROOT_DIR}"/glibc-2.34

sed -i 's/_Static_assert (sizeof (collseqmb) == 256)/_Static_assert (sizeof (collseqmb) == 256, "size of collseqmb")/' \
"${ROOT_DIR}"/backport-Add-codepoint_collation-support-for-LC_COLLATE.patch

pushd "${ROOT_DIR}"
rpmbuild -bp -D "_sourcedir ${ROOT_DIR}" -D "_builddir ${ROOT_DIR}" glibc.spec --nodeps
popd

cp -rf "${ROOT_DIR}"/glibc-2.34/* "${ROOT_DIR}"

cd "${CI_DIR}"
export LD_LIBRARY_PATH=$(echo $LD_LIBRARY_PATH | sed 's/:$//')

"${ROOT_DIR}"/configure --prefix=/ --disable-crypt --enable-bind-now --enable-stack-protector=all \
    CFLAGS="${CFLAGS_ENV}" LDFLAGS="${CFLAGS_ENV}"
make -j64

if [ $? -ne 0 ]; then
    echo "build glibc failed!"
    exit 1
fi

mkdir -p "${CI_DIR}"/output/lib
cp "${CI_DIR}"/libc.so "${CI_DIR}"/output/lib/
cp "${CI_DIR}"/elf/ld.so "${CI_DIR}"/output/lib/
cp "${CI_DIR}"/nptl/libpthread.so "${CI_DIR}"/output/lib/
cp "${CI_DIR}"/dlfcn/libdl.so "${CI_DIR}"/output/lib/
cp "${CI_DIR}"/math/libm.so "${CI_DIR}"/output/lib/

find "${CI_DIR}"/output/lib -type f | xargs -n1 strip -sv

exit 0
