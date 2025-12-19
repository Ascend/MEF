#!/bin/bash
# Perform  build inference
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
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

OUTPUT_NAME="nginx-manager"
DOCKER_FILE_NAME="Dockerfile"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="7.3.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '1p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*=}
fi
arch=$(arch 2>&1)

function clean() {
  rm -rf "${TOP_DIR}/output"
  mkdir -p "${TOP_DIR}/output"
  cd "${TOP_DIR}" && go mod tidy
}

function buildNginx() {
    cd "${TOP_DIR}/cmd"
    export GONOSUMDB="*"
    export CGO_ENABLED=1
    export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    go build -mod=mod -buildmode=pie -ldflags "-s -linkmode=external -extldflags=-Wl,-z,now \
            -X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version}_linux-${arch}" \
            -o ${OUTPUT_NAME} \
            -trimpath
    ls ${OUTPUT_NAME}
    if [ $? -ne 0 ]; then
      echo "fail to find ${OUTPUT_NAME}"
      exit 1
    fi

    cd "${TOP_DIR}"/../opensource/pcre2/
    autoreconf -ivf
    ./configure

    cd "${TOP_DIR}/../opensource/nginx/"
    chmod 750 auto/configure
    CFLAG="-Wall -O2 -fstack-protector-strong -fPIE"
    LDFLAG="-Wl,-z,relro,-z,now,-z,noexecstack -pie -s"


    ./auto/configure --prefix=/home/MEFCenter \
     --with-openssl="${TOP_DIR}"/../opensource/openssl/ \
     --with-pcre="${TOP_DIR}"/../opensource/pcre2/ \
     --with-openssl-opt='-Wall -fPIC -fstack-protector-all -O2 -fomit-frame-pointer' \
     --with-pcre-opt='-Wall -fPIC -fstack-protector-all -O2 -fomit-frame-pointer' \
     --conf-path=/home/MEFCenter/conf/nginx.conf \
     --error-log-path=/home/MEFCenter/logs/error.log \
     --http-log-path=/home/MEFCenter/logs/access.log \
     --pid-path=/home/MEFCenter/logs/nginx.pid \
     --lock-path=/home/MEFCenter/logs/nginx.lock \
     --with-http_ssl_module \
     --http-client-body-temp-path=/tmp/client_body_temp \
     --http-proxy-temp-path=/tmp/proxy_temp \
     --http-fastcgi-temp-path=/tmp/fastcgi_temp \
     --http-uwsgi-temp-path=/tmp/uwsgi_temp \
     --http-scgi-temp-path=/tmp/scgi_temp \
     --with-cc-opt="$CFLAG" --with-ld-opt="$LDFLAG" --without-http_auth_basic_module

    make -j64
    cp "${TOP_DIR}/../opensource/nginx/objs/nginx" "${TOP_DIR}/output/nginx_bin"
}

function mv_file() {
  mkdir -p "${TOP_DIR}/output/nginx/conf"
  cp -R "${TOP_DIR}/../opensource/nginx/conf/mime.types" "${TOP_DIR}/output/nginx/conf/"
  cp "${TOP_DIR}/build/nginx_default.conf" "${TOP_DIR}/output/nginx/conf/"

  mv "${TOP_DIR}/output/nginx_bin" "${TOP_DIR}/output/nginx/nginx"
  mkdir -p "${TOP_DIR}/output/nginx/lib"
  cp "$(realpath "${TOP_DIR}"/..)"/output/lib/libssl.so* "${TOP_DIR}/output/nginx/lib"

  chmod 700 "${TOP_DIR}"/output/nginx/conf
  chmod 400 "${TOP_DIR}"/output/nginx/conf/*
  cp "${TOP_DIR}/cmd/${OUTPUT_NAME}" "${TOP_DIR}/output/nginx/"
  chmod 500 "${TOP_DIR}"/output/nginx/"${OUTPUT_NAME}"
  cp -R "${TOP_DIR}/build/html" "${TOP_DIR}/output/nginx/"
  chmod 700 "${TOP_DIR}"/output/nginx/html
  chmod 400 "${TOP_DIR}"/output/nginx/html/*
  cp "${TOP_DIR}/build/${OUTPUT_NAME}.yaml" "${TOP_DIR}/output/${OUTPUT_NAME}.yaml"
  chmod 600 "${TOP_DIR}"/output/"${OUTPUT_NAME}".yaml
  cp "${TOP_DIR}/build/${DOCKER_FILE_NAME}" "${TOP_DIR}/output/${DOCKER_FILE_NAME}"
  chmod 400 "${TOP_DIR}"/output/"${DOCKER_FILE_NAME}"
  chmod 700 "${TOP_DIR}"/output/nginx

}

function main() {
  clean
  buildNginx
  mv_file
}
main