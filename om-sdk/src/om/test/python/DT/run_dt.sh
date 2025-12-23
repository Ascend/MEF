#!/usr/bin/env bash

set -e

PYTHON3=$(type -p python3)
PYTHON_HOME="${PYTHON3%/*}/"
PIP3=$(type -p pip3)
CUR_PATH=$(cd "$(dirname "$0")";pwd)
echo "CUR_PATH=${CUR_PATH}"
ROOT_PATH=$(readlink -f ${CUR_PATH}/../../..)
echo "ROOT_PATH=${ROOT_PATH}"
SRC_PATH="${ROOT_PATH}/test/python/src/python/"
echo "SRC_PATH=${SRC_PATH}"
TEST_ROOT=${ROOT_PATH}/test/python
echo "TEST_PATH=${TEST_ROOT}"
TEST_SRC_ROOT=${TEST_ROOT}/llt_src
echo "TEST_SRC_ROOT=${TEST_SRC_ROOT}"
LLT_RESULT_ROOT=${TEST_ROOT}/DT/result
echo "LLT_RESULT_ROOT=${LLT_RESULT_ROOT}"
LLT_SETUP_ROOT=${TEST_ROOT}/DT/setup/
echo "LLT_SETUP_ROOT=${LLT_SETUP_ROOT}"
LLT_RESULT_COV_DIR=${LLT_RESULT_ROOT}/cov_data/
echo "LLT_RESULT_COV_DIR=${LLT_RESULT_COV_DIR}"
LLT_RESULT_XML_DIR=${LLT_RESULT_ROOT}/xmls/
echo "LLT_RESULT_XML_DIR=${LLT_RESULT_XML_DIR}"
LLT_RESULT_HTMLS_DIR=${LLT_RESULT_ROOT}/htmls/
echo "LLT_RESULT_HTMLS_DIR=${LLT_RESULT_HTMLS_DIR}"
MCS_LLT_REQUIREMENTS_FILE_PATH=${TEST_ROOT}/DT/requirements.txt
# Collect the number of cases that are executed slowly.
TEST_SLOWEST_NUMBER=5


export OM_WORK_DIR=${ROOT_PATH}/test/install
export OM_WORK_RUN=${ROOT_PATH}/test/install
export OM_MINI_SYSTEM_FLAG_FILE=/home/data/minisys
export PYTHONPATH=${ROOT_PATH}/src/app/sys_om:${ROOT_PATH}/scripts/python:${PYTHONPATH}
export LD_LIBRARY_PATH=/usr/local/mindx/MindXOM/lib:${LD_LIBRARY_PATH}
echo "$OM_WORK_DIR"

function install_om()
{
    echo "install pm pkg start."
    rm -rf ${OM_WORK_DIR}
    mkdir -p ${OM_WORK_DIR}
    if (($? != 0)); then
        echo "[ERROR]mkdir ${OM_WORK_DIR} failed"
        return 1
    fi

    cd ${OM_WORK_DIR}
    unzip -od ${ROOT_PATH}/output ${ROOT_PATH}/output/Ascend-om*.zip
    cp -rf ${ROOT_PATH}/output/Ascend-om*.tar.gz ${OM_WORK_DIR}
    if (($? != 0)); then
        echo "[ERROR]cp -rf ${ROOT_PATH}/output/Ascend-om*.tar.gz ${OM_WORK_DIR} failed"
        return 1
    fi

    tar -zxvf ${OM_WORK_DIR}/Ascend-om*.tar.gz >/dev/null 2>&1
    if (($? != 0)); then
        echo "[ERROR]tar -zxvf ${OM_WORK_DIR}/Ascend-om*.tar.gz"
        return 1
    fi

    # 整合om和omsdk代码
    bash ${OM_WORK_DIR}/scripts/integrate_om_files.sh ${OM_WORK_DIR}  ${OM_WORK_DIR}/om-sdk.tar.gz ${OM_WORK_DIR}/A500-A2-om.tar.gz ${OM_WORK_DIR}/version.xml

    rm -rf ${OM_WORK_DIR}/*.gz*
    if (($? != 0)); then
        echo "[ERROR ]rm -rf ${OM_WORK_DIR}/*.gz*"
        return 1
    fi
    cd -
    echo "install pm pkg finished."
    return 0
}

function safe_replace_os_cmd_conf()
{
    if [ $# -lt 2 ]; then
        echo "safe replace os cmd conf parameter error"
        return 1
    fi
    local src_dirpath="$1"
    local dst_filepath="$2"
    local os_name
    local os_version_id
    local os_cmd_file_list=(
            "os_cmd_eluros2.0.conf"
            "os_cmd_ubuntu22.04.conf"
            "os_cmd_openEuler_22.03.conf"
        )

    # 1、 筛选出源目录下的os_cmd*.conf文件
    for file in "${os_cmd_file_list[@]}"; do
        local file_path="${src_dirpath}/${file}"
        os_name=$(< "${file_path}" grep "OS_NAME" | awk -F "=" '{print $2}' | tr -d '"')
        os_version_id=$(< "${file_path}" grep "OS_VERSION_ID" | awk -F "=" '{print $2}' | tr -d '"')
        echo "os_name is: ${os_name}  os_version_id is: ${os_version_id}"

        # 2、与系统信息进行匹配校验
        if ! check_os_info "${os_name}" "${os_version_id}"; then
            continue
        fi
        echo "os_name is: ${os_name}  os_version_id is: ${os_version_id}"

        # 3、调用拷贝函数进行拷贝
        cp -f "${file_path}" "${dst_filepath}"
        echo "replace os cmd conf from ${file_path} to ${dst_filepath} success"
        return 0
    done

    # 4、如果配置文件匹配失败，则默认将os_cmd_eluros2.0.conf文件内容作为默认的配置
    echo "replace os cmd conf to ${dst_filepath} failed,will replace default"
    cp -f "${src_dirpath}"/os_cmd_eluros2.0.conf "${dst_filepath}"

    return 0
}

function init_env()
{
    install_om
    if (($? != 0)); then
        echo "[ERROR ] install_om failed"
        return 1
    fi

    mkdir -p ${LLT_RESULT_COV_DIR}
    mkdir -p ${LLT_RESULT_XML_DIR}
    mkdir -p ${LLT_RESULT_HTMLS_DIR}

    rm -rf ${LLT_RESULT_COV_DIR}/*
    rm -rf ${LLT_RESULT_XML_DIR}/*
    rm -rf ${LLT_RESULT_HTMLS_DIR}/*

    # 文件拷贝
    mkdir -p /home/data
    mkdir -p /home/data/ies
    mkdir -p /home/data/config
    mkdir -p /home/data/minisys
    cp -rf ${OM_WORK_RUN}/software/ibma/lib/Linux/config/* /home/data/ies/
    source "${OM_WORK_RUN}"/scripts/safe_common.sh
    safe_replace_os_cmd_conf "${OM_WORK_RUN}"/config "/home/data/config/os_cmd.conf"

    cp_so_lib

    mkdir -p /var/plog/ibma_edge/
    echo "init atlas om environment finished."
    return 0
}

function get_coverage_info()
{
    cov_dat_file_num=`ls -l .coverage* |wc -l`
    if [[ ${cov_dat_file_num} -gt 1 ]]; then
        echo "More than one coverage files"
        mv .coverage.`hostname`.*  ${LLT_RESULT_COV_DIR}/.coverage
    else
        mv .coverage ${LLT_RESULT_COV_DIR}
    fi

    ls -la

    cd ${LLT_RESULT_COV_DIR}
    ls -la

    ${PYTHON_HOME}coverage xml
}

function clean_env()
{
     rm -rf ${OM_WORK_DIR}
     echo "rm ${OM_WORK_DIR} finished"
}

function run_tests_of_mgs()
{
    py_test_ret=0
    init_env
    if (($? != 0)); then
        echo "[ERROR ] init_env failed"
        return 1
    fi

    echo "******************before:${PYTHON_HOME}*********"
    ${PIP3} install -r ${MCS_LLT_REQUIREMENTS_FILE_PATH}
    echo "******************end:${PYTHON_HOME}*********"

    cd ${TEST_ROOT}

    local software_path=${OM_WORK_DIR}/software
    export PYTHONPATH=${software_path}/RedfishServer:${software_path}/ibma:${OM_WORK_DIR}/scripts/python:${OM_WORK_DIR}/test/python/llt_src

    ${PYTHON_HOME}pytest ${TEST_SRC_ROOT}/\
    --cov=${OM_WORK_RUN}/software/RedfishServer/om_fd_msg_process \
    --cov=${OM_WORK_RUN}/software/RedfishServer/om_system_service \
    --cov=${OM_WORK_RUN}/software/RedfishServer/upgrade_hdd_service \
    --cov-config="${TEST_SRC_ROOT}"/.coveragerc\
    --junit-xml=${LLT_RESULT_XML_DIR}/final.xml\
    --html=${LLT_RESULT_HTMLS_DIR}/final.html \
    --self-contained-html \
    --durations=${TEST_SLOWEST_NUMBER} \
    --cov-branch \
    --verbose

    py_test_ret=$?
    if ((${py_test_ret} == 0)); then
        echo "Running DT: pytest success, will get_coverage_info"
        get_coverage_info
    fi

    clean_env
    echo "Running DT for OM over. py_test_ret=${py_test_ret}"
    return ${py_test_ret}
}

function cp_so_lib()
{
	  local om_dir="/usr/local/mindx/MindXOM"
	  [ ! -d "${om_dir}" ] && mkdir -p "${om_dir}"
    cp -rf "${OM_WORK_DIR}/lib" "${om_dir}"
    echo "cp lib finished"
}


function main(){
    echo "Running LLT for om now..."
    run_tests_of_mgs
    ret=$?
    echo "All LLT for om over"
    pwd
    return ${ret}
}

start=$(date +%s)
main
main_ret=$?
end=$(date +%s)
echo "LLT running take: $(expr ${end} - ${start}) seconds"
exit ${main_ret}
