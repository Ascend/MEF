#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
LIB_DIR=$(realpath "${TOP_DIR}"/../output/lib)
export GO111MODULE="on"
export GOPATH="/opt/buildtools/go"
export PATH=$GOPATH/bin:$PATH
export LD_LIBRARY_PATH="${LIB_DIR}":"$LD_LIBRARY_PATH"

TEST_MODE=$1

function execute_edge_installer_ut() {
    local build_tag="MEFEdge_A500"
    if [ "$TEST_MODE" == "MEF_Edge_SDK" ]; then
        build_tag="MEFEdge_SDK"
    fi
    echo "test build tag=$build_tag"
    if ! (go test -tags="${build_tag}" -gcflags=-l -v -mod=mod -coverprofile cov.out "${TOP_DIR}"/... >./$file_input); then
        cat ./"$file_input"
        echo '****** edge-installer go test cases error! ******'
        return 1
    fi

    echo "${file_detail_output}"
    gocov convert cov.out >gocov.json
    gocov convert cov.out | gocov-html >"${file_detail_output}"
    gotestsum --junitfile "${ut_xml_output}" -- -tags="${build_tag}" -gcflags=-l "${TOP_DIR}"/...
    return 0
}

file_input='testEdgeInstaller.txt'
file_detail_output='api.html'
DB_PATH="/etc/mindx-edge/edge-installer/"
ut_xml_output="unit-tests.xml"
echo "************************************* Start Edge-Installer LLT Test *************************************"
mkdir -p "${DB_PATH}"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
if [ -f "$file_detail_output" ]; then
    rm -rf $file_detail_output
fi
if [ -f "$file_input" ]; then
    rm -rf "$file_input"
fi
execute_edge_installer_ut
ret=$?
rm -rf "${DB_PATH}"
echo "************************************* End Edge-Installer LLT Test *************************************"

exit $ret
