#!/bin/bash

CUR_DIR=$(dirname "$(readlink -f "$0")")
OM_FILES_LIST=("bin" "config" "lib" "scripts" "software" "tools" "version.xml" "uninstall.sh")
UPGRADE_FINISHED_FLAG="/home/data/upgrade_finished_flag"
OM_UPGRADE_DIR="/usr/local/mindx/MindXOMUpgrade"

source "${CUR_DIR}"/scripts/log_print.sh

init_log_path "/var/plog" "upgrade.log"

function main()
{
    set -e

    # 将软件包的内容拷贝到OM升级目录
    rm -rf "${OM_UPGRADE_DIR}"
    mkdir "${OM_UPGRADE_DIR}"
    chmod 755 "${OM_UPGRADE_DIR}"
    for (( i= 0; i < "${#OM_FILES_LIST[@]}"; i++)); do

        if [[ "${CUR_DIR}" == "/" ]]; then
            logger_error "The installation directory is '/'."
            return 1
        fi

        if ! cp -r "${CUR_DIR}/${OM_FILES_LIST[i]}" "${OM_UPGRADE_DIR}"; then
            logger_error "copy ${OM_FILES_LIST[i]} failed"
            return 1
        fi
    done

    # 执行升级
    export LD_LIBRARY_PATH="${OM_UPGRADE_DIR}"/lib:"${LD_LIBRARY_PATH}"
    export PYTHONPATH="${OM_UPGRADE_DIR}"/software/ibma:"${OM_UPGRADE_DIR}"/software/ibma/opensource/python:"${OM_UPGRADE_DIR}"/scripts/python
    if ! python3 -u "${OM_UPGRADE_DIR}"/scripts/python/upgrade_om.py; then
        logger_error "upgrade omsdk failed"
        return 1
    fi

    # 创建OM升级成功标记
    if ! touch "${UPGRADE_FINISHED_FLAG}"; then
        logger_error "create om upgrade flag failed"
        return 1
    fi

    logger_info "Upgrade MindXOM success."
    return 0
}

main
exit $?