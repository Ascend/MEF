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
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)

TEST_MODE=$1

function call_component_test() {
    echo "************************component($1) test start..."
    export LD_LIBRARY_PATH="${CUR_DIR}/lib":$LD_LIBRARY_PATH

    cd "${TOP_DIR}"/$1/build
    dos2unix test.sh
    chmod +x test.sh
    sh -x test.sh "$TEST_MODE"
    if [[ $? -ne 0 ]]; then
        exit 1
    fi
    sudo cp -rf "${TOP_DIR}"/$1/test/api*.html ${TOP_DIR}/test/results/
    sudo cp -rf "${TOP_DIR}"/$1/test/unit-tests*.xml ${TOP_DIR}/test/results/

    echo "************************component($1) test end. "
}

sudo mkdir -p ${TOP_DIR}/test/results/

sh ${TOP_DIR}/build/build_c_package.sh

echo "************************************* Start MEF_Edge LLT Test *************************************"

call_component_test "edge-installer"

echo "************************************* End MEF_Edge LLT Test *************************************"

exit 0
