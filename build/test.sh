#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2021. All rights reserved.
set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
TEST_SCRIPT="${CUR_DIR}"/test_component.sh

function call_component_test(){
  echo "************************component($1) test start..."

  cd ${TOP_DIR}/$2/build
  work_dir=$(realpath "${TOP_DIR}"/$2)
  sh -x $TEST_SCRIPT $1 $work_dir
  if [[ $? -ne 0 ]]; then
     exit 1
  fi

  results_dir=${TOP_DIR}/test/results
  if [ ! -d "${results_dir}" ]; then
    sudo mkdir -p $results_dir
    sudo cp -rf ${TOP_DIR}/$2/test/api*.html ${TOP_DIR}/test/results/
    sudo cp -rf ${TOP_DIR}/$2/test/unit-tests*.xml ${TOP_DIR}/test/results/
  fi

  echo "************************component($1)  test end. "
}

dos2unix $TEST_SCRIPT
chmod +x $TEST_SCRIPT

echo "************************************* Start MEF_Center LLT Test *************************************"

call_component_test "base" .
call_component_test "edge-manager" ./edge-manager
call_component_test "software-manager" ./software-manager

echo "************************************* End MEF_Center LLT Test *************************************"

exit 0