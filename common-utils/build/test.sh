#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
set -e
export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
export GOFLAGS="-gcflags=all=-l"
unset GOPATH
readonly CUR_DIR=$(dirname "$(readlink -f "$0")")
readonly TOP_DIR=$(realpath "${CUR_DIR}"/..)

readonly components=("hwlog" "utils" "rand" "limiter" "cache" "tls" "x509" "k8stool" "xcrypto" "database"\
    "checker" "modulemgr" "logmgmt" "envutils" "fileutils" "terminal" "websocketmgr"\
    )

readonly TMP_DIR="${TOP_DIR}/test_dir_tmp"
readonly TEST_DIR="${TOP_DIR}/test_dir"

# create test folder for common-utils
function prepare_env_for_common_utils() {
    if [ -d "${TMP_DIR}" ]; then
        rm -rf "${TMP_DIR}"
    fi
    if [ -d "${TEST_DIR}" ]; then
        rm -rf "${TEST_DIR}"
    fi
    mkdir "${TMP_DIR}"
    mkdir "${TEST_DIR}"
}

# create test folder for each component
function prepare_component_env() {
    component=$1
    cd "${TOP_DIR}/${component}"
    if [ -d "test" ]; then
        rm -rf "test"
    fi
    mkdir "test"
}

function execute_test() {
    component=$1
    prepare_component_env "${component}"
    echo "************************************* Start ${component} LLT Test *************************************"
    cd "${TOP_DIR}/${component}/test"
    if ! (go test -mod=mod -v -parallel=1 -gcflags="all=-l -N" -coverprofile "${component}.out" "${TOP_DIR}/${component}/..." \
        >./"${component}.txt"); then
        echo "****** ${component} go test cases error! ******"
        cat "${component}.txt"
        exit 1
    else
        gocov convert "${component}.out" > "${component}.json"
        gocov convert "${component}.out" | gocov-html >"${component}.html"
        gotestsum --junitfile "${component}-unit-tests.xml" -- -mod=mod -v -parallel=1 -gcflags="all=-l -N" "${TOP_DIR}/${component}"/...
        copy_test_file "${component}"
    fi
}

function copy_test_file() {
    component=$1
    cp "${TOP_DIR}/${component}/test/"* "${TMP_DIR}"
}

function merge_llt_result_and_create_new_file() {
    python "${TOP_DIR}/build/llt_result_merge.py" --src="${TMP_DIR}" --det="${TEST_DIR}"
    export PATH=$GOPATH/bin:$PATH
    cat "${TEST_DIR}/gocov.json" | gocov-html > "${TEST_DIR}/api.html"
    cp "${TEST_DIR}"/* "${TOP_DIR}/test/"
}

prepare_env_for_common_utils

for component in "${components[@]}"
do
  execute_test "$component"
done

merge_llt_result_and_create_new_file
