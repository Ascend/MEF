#!/bin/bash
# Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
declare -i ret_ok=0  # success

CURRENT_PATH=$(dirname "$(readlink -f "$0")")
cd "${CURRENT_PATH}" || exit 1
TOOL_BINARY_PATH="${CURRENT_PATH}/bin"
UPGRADE_SH_PATH="${CURRENT_PATH}/upgrade.sh"

print_helps()
{
    echo "MEF run entry cmd help:"
    echo "--------control--------- ./run.sh start edge_manager "
    echo "--------upgrade--------- ./run.sh upgrade "
    echo "--------uninstall--------- ./run.sh uninstall"

}

function main()
{
    if [[ "$#" -gt 0 ]];  then
        method="$1"
        binary_file=""
        case "$method" in
             "start" | "stop" | "restart")
                binary_file="MEF-center-controller"
                manage_type="operate"
                if [[ -z $2 ]]; then
                    component="all"
                else
                    component="$2"
                fi

                params="-$method=$component"
            ;;
             "uninstall")
                binary_file="MEF-center-controller"
                manage_type="uninstall"
            ;;
            "-h" | "--help" | "--h")
                print_helps "$@"
                exit 0
            ;;
            * )
                echo "The input params not valid; please read help and try again"
                print_helps "$@"
                exit 1
            ;;
        esac

        if [[ "${manage_type}" == "operate" ]]; then
            "${TOOL_BINARY_PATH}"/"${binary_file}" "$params" "-operate=operate"

            ret=$?
            if [[ "${ret}" != 0 ]];then
                echo "${method} ${component} component failed"
                return ${ret}
            fi
            echo "${method} ${component} component success"
        else
            "${TOOL_BINARY_PATH}"/"${binary_file}" "-operate=$manage_type"

            ret=$?
            if [[ "${ret}" != 0 ]];then
                echo "${method} MEF Center failed"
                return ${ret}
            fi
            echo "${method} MEF Center success"
        fi

        return ${ret_ok}
    fi
}

main "$@"
RESULT=$?
exit ${RESULT}