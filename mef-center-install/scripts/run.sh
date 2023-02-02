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
    echo "-version -- display the program name and version"
    echo "-h/--help/--h -- display the helps information"
    echo "start -- start all or a component"
    echo "parameters: "
    echo "        all: start all component, default value"
    echo "        {component_name}: start a single component. eg: run.sh start edge-manager"
    echo " "
    echo "stop -- stop all or a component"
    echo "parameters: "
    echo "        all: stop all component, default value"
    echo "        {component_name}: stop a single component. eg: run.sh stop edge-manager"
    echo " "
    echo "restart -- restart all or a component"
    echo "parameters: "
    echo "        all: restart all component, default value"
    echo "        {component_name}: restart a single component. eg: run.sh restart edge-manager"
    echo " "
    echo "uninstall -- uninstall MEF Center"
}

function main()
{
    binary_file="MEF-center-controller"
    if [[ "$#" -gt 0 ]];  then
        method="$1"
        case "$method" in
             "start" | "stop" | "restart")
                manage_type="operate"
                if [[ -z $2 ]]; then
                    component="all"
                else
                    component="$2"
                fi

                params="-$method=$component"
            ;;
             "uninstall")
                manage_type="uninstall"
            ;;
            "-h" | "--help" | "--h")
                print_helps "$@"
                exit 0
            ;;
            "-version" )
                "${TOOL_BINARY_PATH}"/"${binary_file}" "-version"
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