#!/bin/bash

set -e

CURRENT_PATH="$(dirname $(readlink -f $0))"
CUR_PATH=$(cd "$(dirname "$0")";pwd)
TOP_DIR=$(readlink -f ${CUR_PATH}/../../..)
OM_INSTALL_PATH=${TOP_DIR}/test/install

function install_om()
{
    echo "install pm pkg start."
    mkdir -p ${OM_INSTALL_PATH}
    if (($? != 0)); then
        echo "[ERROR]mkdir ${OM_INSTALL_PATH} failed"
        return 1
    fi

    cd ${OM_INSTALL_PATH}
    unzip -od ${TOP_DIR}/output ${TOP_DIR}/output/Ascend-mindxedge-omsdk*.zip
    cp -f ${TOP_DIR}/output/Ascend-mindxedge-omsdk*.tar.gz ${OM_INSTALL_PATH}
    if (($? != 0)); then
        echo "[ERROR]cp -f ${TOP_DIR}/output/Ascend-mindxedge-omsdk*.tar.gz ${OM_INSTALL_PATH} failed"
        return 1
    fi

    tar -zxvf ${OM_INSTALL_PATH}/Ascend-mindxedge-omsdk*.tar.gz >/dev/null 2>&1
    if (($? != 0)); then
        echo "[ERROR]tar -zxvf ${OM_INSTALL_PATH}/Ascend-mindxedge-omsdk*.tar.gz"
        return 1
    fi

    rm -rf ${OM_INSTALL_PATH}/*.gz*
    if (($? != 0)); then
        echo "[ERROR ]rm -rf ${OM_INSTALL_PATH}/*.gz*"
        return 1
    fi

    cd -
    echo "install pm pkg finished."
    return 0
}

function build_prepare()
{
    install_om
    if (($? != 0)); then
        echo "[ERROR ]install_om failed"
        return 1
    fi

    mkdir -p ${TOP_DIR}/test/C++/output
    mkdir -p ${TOP_DIR}/test/C++/lib/
    mkdir -p ${TOP_DIR}/test/C++/src/build

    cp -rf ${TOP_DIR}/opensource/libboundscheck/include/* ${TOP_DIR}/test/C++/output
    cp -rf ${TOP_DIR}/opensource/googletest/googletest/include/* ${TOP_DIR}/test/C++/output
    mkdir -p ${TOP_DIR}/test/C++/output/crypto/rsa/
    cp -f ${TOP_DIR}/opensource/openssl/crypto/rsa/rsa_local.h "${TOP_DIR}"/test/C++/output/crypto/rsa/
    mkdir -p ${TOP_DIR}/test/C++/output/crypto/evp/
    cp -f "${TOP_DIR}"/opensource/openssl/crypto/evp/evp_local.h "${TOP_DIR}"/test/C++/output/crypto/evp/

    cp -rf ${TOP_DIR}/test/install/lib/libcrypto.so* ${TOP_DIR}/test/C++/lib/
    cp -rf ${TOP_DIR}/test/install/lib/libssl.so* ${TOP_DIR}/test/C++/lib/
    cp -rf ${TOP_DIR}/test/install/lib/libsecurec.so ${TOP_DIR}/test/C++/lib/

    mkdir -p /home/data/config/default/
    # DT用例测试需要在执行前创建日志目录(生产环境执行是由om_init.sh脚本创建)，和删除上次DT执行产生的日志
    mkdir -p /var/plog/ibma_edge/
    rm -rf /var/plog/ibma_edge/om_scripts_run.log
    return 0
}


function start_build_ut()
{
    echo "start run test..."
    pushd ${TOP_DIR}/test/C++/src/build/
         cmake ..
         make clean
         make
         ret=$?
    popd
    return $ret
}

function start_run_ut()
{
    export ENS_HOME=${TOP_DIR}/test/install/software/ens
    export LD_LIBRARY_PATH=${TOP_DIR}/test/C++/lib:$LD_LIBRARY_PATH
    local ret=0
    pushd ${TOP_DIR}/test/C++/output/
        ./MindXOM_TEST --gtest_output=xml:test_detail.xml
        ret=$?
    popd
    return $ret
}

function start_generate_coverage()
{
    local coverage_file="$TOP_DIR"/test/C++/output/coverage.info
    local report_folder="$TOP_DIR"/test/C++/result
    pushd ${TOP_DIR}
    lcov --rc lcov_branch_coverage=1 -c -d . -o "${coverage_file}" \
        --exclude "${TOP_DIR}/opensource/*" \
        --exclude "${TOP_DIR}/test/*" \
        --exclude "/usr/include/*"
    genhtml --rc genhtml_branch_coverage=1 "${coverage_file}" -o "${report_folder}"
    popd
}

function clear_env()
{
    rm -rf ${OM_INSTALL_PATH}
    return 0
}

function main()
{
    echo "build prepare."
    build_prepare
    if [ $? -ne 0 ]; then
        echo "run ut failed for build_prepare!"
        clear_env
        return 1
    fi

    start_build_ut
    if [ $? -ne 0 ]; then
        echo "build ut failed for start_build_ut!"
        clear_env
        return 1
    fi
    echo "build ut success!"

    start_run_ut
    if [ $? -ne 0 ]; then
        echo "run ut failed for start_run_ut!"
        clear_env
        return 1
    fi
    echo "run ut success!"

    start_generate_coverage
    if [ $? -ne 0 ]; then
        echo "generate coverage failed for start_generate_coverage!"
        clear_env
        return 1
    fi
    echo "generate coverage success!"

    clear_env
    return 0
}

main
ret=$?
echo "test finished with ret $ret"
exit $ret
