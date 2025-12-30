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
set -e

# ../atlas-base/build
CUR_DIR=$(dirname "$(readlink -f $0)")
# ../atlas-base
TOP_DIR=$(realpath "${CUR_DIR}"/..)
TMP_DIR="${TOP_DIR}/test_tmp"
TEST_DIR="${TOP_DIR}/test/results"

echo "TOP_DIR=${TOP_DIR}"
EDGE_MANAGER="edge-manager"
ALARM_MANAGER="alarm-manager"
CERT_MANAGER="cert-manager"

function clean() {
  if [ -f "${TMP_DIR}" ]; then
    rm -rf "${TMP_DIR}"
  fi
  if [ -f "${TEST_DIR}" ]; then
    rm -rf "${TEST_DIR}"
  fi
  mkdir -p "${TMP_DIR}"
  mkdir -p "${TEST_DIR}"
}

function execute_edge_manager_ut() {
  cd "${TOP_DIR}"/edge-manager/build/
  chmod u+x test.sh
  ./test.sh
}

function execute_alarm_manager_ut() {
  cd "${TOP_DIR}"/alarm-manager/build/
  chmod u+x test.sh
  ./test.sh
}

function execute_cert_manager_ut() {
  cd "${TOP_DIR}"/cert-manager/build/
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
  copy_file ${ALARM_MANAGER}
  copy_file ${CERT_MANAGER}
}

function merge_llt_result_and_create_new_file() {
    python "${TOP_DIR}/build/llt_result_merge.py" --src="${TMP_DIR}" --det="${TEST_DIR}"
    export PATH=$GOPATH/bin:$PATH
    cat "${TEST_DIR}/gocov.json" | gocov-html > "${TEST_DIR}/api.html"
}

function main() {
  clean
  execute_alarm_manager_ut
  execute_cert_manager_ut
  execute_edge_manager_ut
  execute_base_ut
  copy_base_file
  copy_llt_result_files
  merge_llt_result_and_create_new_file
}

main