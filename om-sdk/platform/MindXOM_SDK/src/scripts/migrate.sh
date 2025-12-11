#!/bin/bash

OM_UPGRADE_DIR="/usr/local/mindx/MindXOMUpgrade"
REDFISH_MIGRATE_ENTRY="${OM_UPGRADE_DIR}/software/RedfishServer/redfish_db/migrate.sh"
MONITOR_MIGRATE_ENTRY="${OM_UPGRADE_DIR}/software/ibma/monitor_db/migrate.py"

# init log
source "${OM_UPGRADE_DIR}"/scripts/log_print.sh
init_log_path "/var/plog" "upgrade.log"

function main()
{
    # 检查当前操作用户是否为root
    if [[ $(whoami) != "root" ]]; then
        logger_error "Current user is not root, please check."
        return 1
    fi

    local operate="$1"
    if [[ "${operate}" != "install" ]] && [[ "${operate}" != "migrate" ]]; then
        logger_error "error:migrate failed: not supported operate type."
        return 1
    fi

    # 执行Monitor数据迁移
    logger_info "start migrate monitor_edge.db"
    (
        export LD_LIBRARY_PATH="${OM_UPGRADE_DIR}"/lib:"${LD_LIBRARY_PATH}"
        export PYTHONPATH="${OM_UPGRADE_DIR}"/software/ibma:"${OM_UPGRADE_DIR}"/software/ibma/opensource/python
        python3 "${MONITOR_MIGRATE_ENTRY}" "${operate}"
    )
    local ret=$?
    if [ "${ret}" -ne 0 ]; then
        logger_error "migrate monitor_edge.db failed"
        return 1
    fi
    logger_info "migrate monitor_edge.db finish."

    # 执行redfish数据库迁移
    logger_info "start migrate redfish_edge.db"
    su - MindXOM -s "$(which bash)" -c "/bin/bash ${REDFISH_MIGRATE_ENTRY} ${operate}"
    ret=$?
    if [ "${ret}" -ne 0 ]; then
        logger_error "migrate redfish_edge.db failed"
        return 1
    fi
    logger_info "migrate redfish_edge.db finish."
    return 0
}

main "$@"
exit $?
