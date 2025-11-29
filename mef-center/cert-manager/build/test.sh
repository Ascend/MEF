#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2023. All rights reserved.
set -e
CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
echo "$TOP_DIR"
export GO111MODULE="on"
export GOPATH="/opt/buildtools/go"
export PATH=$GOPATH/bin:$PATH

function execute_edge_manager_ut() {
  if ! (go test -tags=TESTCODE -gcflags=-l -v -mod=mod -coverprofile cov.out "${TOP_DIR}/..." >"./$file_input"); then
    echo '****** cert-manager go test cases error! ******'
    cat "$file_input"
    exit 1
  else
    echo "${file_detail_output}"
    gocov convert cov.out > gocov.json
    gocov convert cov.out | gocov-html > "${file_detail_output}"
    gotestsum --junitfile unit-tests.xml -- -tags=TESTCODE -gcflags=-l "${TOP_DIR}"/...
  fi
}

file_input='testEdgeManager.txt'
file_detail_output='api.html'
DB_PATH="/home/data/config/"

echo "************************************* Start cert-manager LLT Test *************************************"
mkdir -p "${TOP_DIR}"/inner-test/
mkdir -p "${DB_PATH}"
cd "${TOP_DIR}"/inner-test/
if [ -f "$file_detail_output" ]; then
  rm -rf $file_detail_output
fi
if [ -f "$file_input" ]; then
  rm -rf $file_input
fi
execute_edge_manager_ut
rm -rf "${DB_PATH}"
echo "************************************* End cert-manager LLT Test *************************************"

exit 0
