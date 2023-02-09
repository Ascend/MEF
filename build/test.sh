#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2021. All rights reserved.
set -e

# ../atlas-base/build
CUR_DIR=$(dirname "$(readlink -f $0)")
# ../atlas-base
TOP_DIR=$(realpath "${CUR_DIR}"/..)
TMP_DIR="${TOP_DIR}/test_tmp"
TEST_DIR="${TOP_DIR}/test/results"

echo "TOP_DIR=${TOP_DIR}"
EDGE_MANAGER="edge-manager"

function prepare(){
  local lib_dir
  lib_dir=$(dirname "${TOP_DIR}")
  echo "lib_dir=$lib_dir"
  tar -zxf "$(ls "$lib_dir"/*kmc*.tar.gz)" -C "$lib_dir"
  sleep 2
  echo "ls_lib=$(ls -l $lib_dir/lib)"
  export LD_LIBRARY_PATH=$lib_dir/lib:"$LD_LIBRARY_PATH"
  echo "LD_LIBRARY_PATH=$LD_LIBRARY_PATH"
}

function clean() {
  rm -rf "${TMP_DIR}"
  rm -rf "${TEST_DIR}"
  mkdir -p "${TMP_DIR}"
  mkdir -p "${TEST_DIR}"
}

function execute_edge_manager_ut() {
  cd "${TOP_DIR}"/edge-manager/build/
  chmod u+x test.sh
  ./test.sh
}

function execute_base_ut() {
  cd "${TOP_DIR}"/build/
  chmod u+x test_component.sh
  ./test_component.sh
}

function copy_file() {
  component=$1
  cp "${TOP_DIR}/${component}/inner-test/gocov.json" "${TMP_DIR}/${component}.json"
  cp "${TOP_DIR}/${component}/inner-test/unit-tests.xml" "${TMP_DIR}/${component}.xml"
}

function copy_base_file() {
  cp "${TOP_DIR}/inner-test/gocov.json" "${TMP_DIR}/base.json"
  cp "${TOP_DIR}/${component}/inner-test/unit-tests.xml" "${TMP_DIR}/base.xml"
}

function copy_llt_result_files() {
  copy_file ${EDGE_MANAGER}
}

function merge_llt_result_and_create_new_file() {
    python "${TOP_DIR}/build/llt_result_merge.py" --src="${TMP_DIR}" --det="${TEST_DIR}"
    export PATH=$GOPATH/bin:$PATH
    cat "${TEST_DIR}/gocov.json" | /opt/buildtools/go/bin/gocov-html > "${TEST_DIR}/api.html"
}

function main() {
  clean
  prepare
  execute_edge_manager_ut
  execute_base_ut
  copy_base_file
  copy_llt_result_files
  merge_llt_result_and_create_new_file
}

main