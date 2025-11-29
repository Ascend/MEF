#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
umask 077
declare -i ret_ok=0  # success

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
export LD_LIBRARY_PATH=${CURRENT_PATH}/lib/kmc-lib:$LD_LIBRARY_PATH
INSTALL_PROGRAM_PATH="${CURRENT_PATH}/bin"
INSTALL_BIN="MEF-center-installer"

function main()
{
    "${INSTALL_PROGRAM_PATH}"/"${INSTALL_BIN}" "$@"
    ret=$?
    # help/version时不打印安装失败信息
    if [[ "${ret}" == 3 ]];then
        return ${ret}
    elif [[ "${ret}" != 0 ]];then
        echo "install MEF center failed"
        return ${ret}
    fi

    echo "install MEF center success"
    return ${ret_ok}
}

main "$@"
RESULT=$?
exit ${RESULT}