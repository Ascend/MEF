#!/bin/bash
CURRENT_PATH="$(dirname $(readlink -f $0))"
CUR_PATH=$(cd "$(dirname "$0")";pwd)
TOP_DIR=$(readlink -f ${CUR_PATH}/../../..)
OM_INSTALL_PATH=${TOP_DIR}/test/install

function install_om()
{
    echo "install pm pkg start."
    sudo mkdir -p ${OM_INSTALL_PATH}
    if (($? != 0)); then
        echo "[ERROR]mkdir ${OM_INSTALL_PATH} failed"
        return 1
    fi

    cd ${OM_INSTALL_PATH}
    sudo unzip -d ${TOP_DIR}/output ${TOP_DIR}/output/Ascend-mindxedge-omsdk*.zip
    sudo cp -f ${TOP_DIR}/output/Ascend-mindxedge-omsdk*.tar.gz ${OM_INSTALL_PATH}
    if (($? != 0)); then
        echo "[ERROR]sudo cp -f ${TOP_DIR}/output/Ascend-mindxedge-omsdk*.tar.gz ${OM_INSTALL_PATH} failed"
        return 1
    fi

    sudo tar -zxvf ${OM_INSTALL_PATH}/Ascend-mindxedge-omsdk*.tar.gz >/dev/null 2>&1
    if (($? != 0)); then
        echo "[ERROR]tar -zxvf ${OM_INSTALL_PATH}/Ascend-mindxedge-omsdk*.tar.gz"
        return 1
    fi

    sudo rm -rf ${OM_INSTALL_PATH}/*.gz*
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

    sudo mkdir -p ${TOP_DIR}/test/C++/output/
    sudo mkdir -p ${TOP_DIR}/test/C++/lib/
    sudo mkdir -p ${TOP_DIR}/test/C++/src/build/

    sudo cp -rf ${TOP_DIR}/output/lib/* ${TOP_DIR}/test/C++/lib/
    sudo cp -rf ${OM_INSTALL_PATH}/software/ens/modules/* ${TOP_DIR}/test/C++/lib/
    sudo cp -rf ${OM_INSTALL_PATH}/software/ens/lib/* ${TOP_DIR}/test/C++/lib/
    sudo chmod 755 ${TOP_DIR}/test/C++/lib/ -R

    # DT用例测试需要在执行前创建日志目录(生产环境执行是由om_init.sh脚本创建)，和删除上次DT执行产生的日志
    sudo mkdir -p /var/plog/ibma_edge/
    sudo rm -rf /var/plog/ibma_edge/om_scripts_run.log
    return 0
}


function start_build_ut()
{
    echo "start build test..."
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
        ./MindXOM_TEST
        ret=$?
    popd
    return $ret
}

function clear_env()
{
    sudo rm -rf ${OM_INSTALL_PATH}
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
    clear_env
    return 0
}

main
ret=$?
echo "test finished with ret $ret"
exit $ret
