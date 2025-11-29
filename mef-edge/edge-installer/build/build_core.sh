#!/bin/bash
# Perform  build inference
# Copyright (c) 2023. Huawei Technologies Co., Ltd. All rights reserved.
set -e
set -x
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_MEF_DIR=$(realpath "${CUR_DIR}"/../..)


export GO111MODULE="on"
OUTPUT_NAME="edgecore"
arch=$(arch 2>&1)
echo "Build Architecture is" "${arch}"

function build() {
  cd "${TOP_MEF_DIR}/opensource/kubeedge"
  if [ -e ${OUTPUT_NAME} ]; then
    echo "edgecore exist, no need to build"
    return
  fi

  export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
  go mod tidy
  go mod vendor
  rm -rf go.sum
  # remove testing rsa key from edgecore
  clear_test_file
  go build -mod=vendor -buildmode=pie -trimpath -ldflags "-buildid=IdAtlasEdge -s -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
          -o "${OUTPUT_NAME}" \
          ./edge/cmd/edgecore/edgecore.go
  ls "${OUTPUT_NAME}"
  if [ $? -ne 0 ]; then
    echo "fail to find ${OUTPUT_NAME}"
    exit 1
  fi

  # raise error if we find a rsa private key in binary
  local num_keylines=$(grep -a "RSA TESTING KEY" "${OUTPUT_NAME}" | wc -l)
  if [ "$num_keylines" -ne 0 ]; then
    echo "unexpected rsa private key found in binary ${OUTPUT_NAME}"
    exit 1
  fi
}

function clear_test_file(){
    local file_name
    local clear_file_list=("runtime_mock.go" "mock_runtime_cache.go" "mock_cni.go" "cadvisor_mock.go" "mock_manager.go" \
    "mock_pod_status_provider.go" "mock_stats_provider.go" "mock_volume.go")
    local vendor_path="vendor"
    echo "vendor_path=$vendor_path"
    for file_name in ${clear_file_list[@]}; do
        file_path=$(find "$vendor_path" -name "$file_name" || true)
        if [ -f "$file_path" ];then
            echo "mock_file_path=$file_path"
            rm -f "$file_path"
        fi
    done
}

function mv_file() {
  mv "${TOP_MEF_DIR}/opensource/kubeedge/${OUTPUT_NAME}" "${TOP_MEF_DIR}/output/"
  chmod 500 "${TOP_MEF_DIR}/output/${OUTPUT_NAME}"
}
function main() {
  build
  mv_file
}
main