#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2021. All rights reserved.
set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)

function call_component_test(){
  echo "************************component($1) test start..."

  cd ${TOP_DIR}/$1/build
  dos2unix test.sh
  chmod +x test.sh
  sh test.sh
  if [[ $? -ne 0 ]]; then
     exit 1
  fi
  sudo cp -rf ${TOP_DIR}/$1/test/api*.html ${TOP_DIR}/test/results/
  sudo cp -rf ${TOP_DIR}/$1/test/unit-tests*.xml ${TOP_DIR}/test/results/

  echo "************************component($1) test end."
}

sudo mkdir -p ${TOP_DIR}/test/results/

echo "************************************* Start MEF_Center LLT Test *************************************"

call_component_test "edge-manager"

echo "************************************* End MEF_Center LLT Test *************************************"

exit 0
