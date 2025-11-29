#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
umask 077
declare -i ret_ok=0  # success
declare -i ret_wrong_arch=2 # no arch ret
declare -i ret_grep_wrong=3 # grep failed
declare -i ret_uname_wrong=4 # uname -i failed

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
UPGRADE_BIN_PATH="${CURRENT_PATH}/../bin"
VERSION_XML_PATH="${CURRENT_PATH}/../version.xml"
UPGRADE_BIN="MEF-center-upgrade"

export LD_LIBRARY_PATH=${CURRENT_PATH}/../lib/kmc-lib:${CURRENT_PATH}/../lib/lib:$LD_LIBRARY_PATH

function check_arch()
{
    arch=$(grep '<ProcessorArchitecture>' "$VERSION_XML_PATH")
    if [ $? != 0 ]; then
      return "${ret_grep_wrong}"
    fi
    arch=${arch#*>}
    arch=${arch%<*}
    ret=$(uname -i)
    if [ $? != 0 ]; then
      return "${ret_uname_wrong}"
    fi
    if [ "$ret" != "$arch" ];then
        return "${ret_wrong_arch}"
    fi
    return "${ret_ok}"
}

function main()
{
    check_arch
    ret=$?
    case "${ret}" in
        "${ret_wrong_arch}")
           echo "invalid arch" >&2
           return "${ret_wrong_arch}"
        ;;
        "${ret_grep_wrong}")
           return "${ret_grep_wrong}"
        ;;
        "${ret_uname_wrong}")
           return "${ret_uname_wrong}"
        ;;
    esac
    "${UPGRADE_BIN_PATH}"/"${UPGRADE_BIN}"
    ret=$?
    return "${ret}"
}

main "$@"
RESULT=$?
exit ${RESULT}