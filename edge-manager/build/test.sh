#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2021. All rights reserved.
set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
export GOPATH="/opt/buildtools/go"
export PATH=$GOPATH/bin:$PATH

function execute_edge_manager_ut() {
  if ! (go test -gcflags=-l -v -mod=mod -coverprofile cov.out ${TOP_DIR}/... >./$file_input); then
    echo '****** edge-manager go test cases error! ******'
    exit 1
  else
    echo ${file_detail_output}
    gocov convert cov.out > gocov.json
    gocov convert cov.out | gocov-html >${file_detail_output}
    gotestsum --junitfile ${ut_xml_output} -- -gcflags=-l "${TOP_DIR}"/...
  fi
}

file_input='testEdgeManager.txt'
file_detail_output='api.html'
DB_PATH="/etc/mindx-edge/edge-manager/"
ut_xml_output="unit-tests.xml"
echo "************************************* Start Edge-Manager LLT Test *************************************"
echo "to delete, current dir:$CUR_DIR"
mkdir -p "${DB_PATH}"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
if [ -f "$file_detail_output" ]; then
  rm -rf $file_detail_output
fi
if [ -f "$file_input" ]; then
  rm -rf $file_input
fi
execute_edge_manager_ut
rm -rf "${DB_PATH}"
echo "************************************* End Edge-Manager LLT Test *************************************"

exit 0
