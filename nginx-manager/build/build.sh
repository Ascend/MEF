#!/bin/bash
# Perform  build inference
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

OUTPUT_NAME="nginx-manager"
DOCKER_FILE_NAME="Dockerfile"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="3.0.0"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '6p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi

arch=$(arch 2>&1)

function buildNginx() {
    cd "${TOP_DIR}/../opensource/nginx/"
    chmod 750 auto/configure
    ./auto/configure --prefix=/home/MEFCenter --conf-path=/home/MEFCenter/conf/nginx.conf --error-log-path=/home/MEFCenter/log/error.log --http-log-path=/home/MEFCenter/log/access.log --pid-path=/home/MEFCenter/log/nginx.pid --lock-path=/home/MEFCenter/log/nginx.lock --with-http_ssl_module --http-client-body-temp-path=/tmp/client_body_temp  --http-proxy-temp-path=/tmp/proxy_temp --http-fastcgi-temp-path=/tmp/fastcgi_temp --http-uwsgi-temp-path=/tmp/uwsgi_temp --http-scgi-temp-path=/tmp/scgi_temp
    make
    cp "${TOP_DIR}/../opensource/nginx/objs/nginx" "${TOP_DIR}/output/nginx_bin"
}

function mv_file() {
  mkdir -p "${TOP_DIR}/output/nginx/dist"
  mkdir -p "${TOP_DIR}/output/nginx/conf"
  cp -R "${TOP_DIR}/../opensource/nginx/conf/mime.types" "${TOP_DIR}/output/nginx/conf/"
  cp "${TOP_DIR}/build/nginx_default.conf" "${TOP_DIR}/output/nginx/conf/"

  mv "${TOP_DIR}/output/nginx_bin" "${TOP_DIR}/output/nginx/nginx"
  cp "${TOP_DIR}/cmd/${OUTPUT_NAME}" "${TOP_DIR}/output/nginx/"
  cp "${TOP_DIR}/build/${OUTPUT_NAME}.yaml" "${TOP_DIR}/output/${OUTPUT_NAME}.yaml"
  cp "${TOP_DIR}/build/${DOCKER_FILE_NAME}" "${TOP_DIR}/output/${DOCKER_FILE_NAME}"

  cd "${TOP_DIR}/output/"
}

function main() {
  buildNginx
  mv_file
}
main