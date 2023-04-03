#!/bin/bash
# Perform  build inference
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

OUTPUT_NAME="nginx-manager"
DOCKER_FILE_NAME="Dockerfile"
VER_FILE="${TOP_DIR}"/service_config.ini
build_version="5.0.0.RC1"
if [ -f "$VER_FILE" ]; then
  line=$(sed -n '6p' "$VER_FILE" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi
arch=$(arch 2>&1)
function mv_file() {

  mkdir -p "${TOP_DIR}/output/nginx/conf"
  cp -R "${TOP_DIR}/../opensource/nginx/conf/mime.types" "${TOP_DIR}/output/nginx/conf/"
  cp "${TOP_DIR}/build/nginx_default.conf" "${TOP_DIR}/output/nginx/conf/"
  cp "${TOP_DIR}/build/nginx_resolver.conf" "${TOP_DIR}/output/nginx/conf/"
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

  cd "${TOP_DIR}/output/"
}

function main() {
  mv_file
}
main