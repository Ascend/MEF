#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
umask 077
declare -i ret_ok=0  # success

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
UPGRADE_BIN_PATH="${CURRENT_PATH}/../bin"
VERSION_XML_PATH="${CURRENT_PATH}/../version.xml"
UPGRADE_BIN="MEF-center-upgrade"

function check_arch()
{
    arch=$(grep '<ProcessorArchitecture>' "$VERSION_XML_PATH")
    arch=${arch#*>}
    arch=${arch%<*}
    ret=$(uname -i)
    if [ "$ret" != "$arch" ];then
        return 1
    fi
    return 0
}

function main()
{
    if ! check_arch; then
       echo "invalid arch" >&2
       return 1
    fi
    "${UPGRADE_BIN_PATH}"/"${UPGRADE_BIN}"
    ret=$?
    return ${ret}
}

main "$@"
RESULT=$?
exit ${RESULT}