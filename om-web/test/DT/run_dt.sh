#!/usr/bin/env bash
CUR_PATH=$(cd "$(dirname "$0")" || exit 1;pwd)
echo "CUR_PATH=${CUR_PATH}"
ROOT_PATH=$(readlink -f "${CUR_PATH}/../../")
COV_PATH="${ROOT_PATH}/test/coverage"


function prepare_env()
{
    npm config set registry https://cmc.centralrepo.rnd.huawei.com/artifactory/api/npm/npm-central-repo/
    npm ci
}

function get_metric()
{
    local clover_xml="${COV_PATH}/clover.xml"
    if [ ! -e "${clover_xml}" ]; then
        echo "${clover_xml} not found"
        return 1
    fi
    local key=$1
    val=$(grep metrics "${clover_xml}" | head -n1 | awk -F "${key}=\"" '{print $2}' | awk -F '"' '{print $1}')
    echo "$val"
}

function set_val()
{
    local fn=$1
    local key=$2
    local val=$3

    sed -i "s/${key}/${val}/g" "$fn"
}

function gen_cov_report()
{
    class_covered=0
    class_missed=0
    method_count=$(get_metric "methods")
    method_covered=$(get_metric "coveredmethods")
    method_missed=$((method_count - method_covered))

    line_count=$(get_metric "statements")
    line_covered=$(get_metric "coveredstatements")
    line_missed=$((line_count - line_covered))

    branch_count=$(get_metric "conditionals")
    branch_covered=$(get_metric "coveredconditionals")
    branch_missed=$((branch_count - branch_covered))

    local js_cov="${ROOT_PATH}/test/reports/js_cov.xml"
    if [ -e "${js_cov}" ];then
        rm "${js_cov}"
    fi
    cp "${ROOT_PATH}/test/DT/js_cov.tpl.xml" "$js_cov"
    set_val "$js_cov" "{class_covered}" "${class_covered}"
    set_val "$js_cov" "{class_missed}" "${class_missed}"
    set_val "$js_cov" "{method_covered}" "${method_covered}"
    set_val "$js_cov" "{method_missed}" "${method_missed}"
    set_val "$js_cov" "{line_covered}" "${line_covered}"
    set_val "$js_cov" "{line_missed}" "${line_missed}"
    set_val "$js_cov" "{branch_covered}" "${branch_covered}"
    set_val "$js_cov" "{branch_missed}" "${branch_missed}"

    start_time=$(date +%s)
    set_val "$js_cov" "{start_time}" "${start_time}"
}

function main() {
    cd "${ROOT_PATH}" || exit 1
    prepare_env

    npm run coverage
    test_ret=$?
    if [ $test_ret -ne 0 ];then
        echo "run test failed"
        return 1
    fi

    gen_cov_report
    gen_cov_ret=$?
    if [ $gen_cov_ret -ne 0 ];then
        echo "gen cov report failed"
        return 1
    fi
}

main
exit $?
