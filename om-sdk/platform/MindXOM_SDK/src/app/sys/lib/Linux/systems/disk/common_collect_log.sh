#! /bin/bash

function check_link()
{
    local file_path="$1"
    local real_path

    real_path=$(readlink -f "${file_path}")
    if [[ "${file_path}" != "${real_path}" ]]; then
        logger_error "file path is softlink!"
        return 1
    fi

    return 0
}

#collect npu log
function npu_log()
{
    if [ ! -d "/var/alog" ]; then
        logger_warn "can not find alog in var"
    fi

    if [ ! -d "/var/plog" ]; then
        logger_warn "can not find plog in var"
    fi

    if [ ! -d "/home/log/plog" ]; then
        logger_warn "can not find plog in home"
    fi

    if [ ! -d "/var/log" ]; then
        logger_warn "can not find log in var"
    fi

    tar --ignore-failed-read -czPf "${LOG_PATH}"/NPU.tar.gz \
     /var/log/ascend_seclog/ascend_install.log \
     /var/log/ascend_seclog/ascend_run_servers.log \
     /var/log/ascend_seclog/operation.log \
     /var/log/device_sys_init.log \
     /var/log/nputools_LOG_*.log \
     /var/log/npu/hisi_logs/ \
     /var/log/kbox \
     /var/log/npu \
     /home/log/dlog \
     /home/log/plog/nputools_LOG_*
    logger_info "Collect npu logs Successfully"
    return 0
}

#collect AtlasEdge log
function MEF_log()
{
    local module=$1
    if [ -z ${module} ];then
        logger_error "MEF module is null"
        return 1
    fi
    local edge_service_path="/usr/lib/systemd/system/mef-edge-main.service"
    if ! check_link "${edge_service_path}";then
        logger_error "MEF service path is invalid"
        return 1
    fi

    local edge_install_path
    if ! edge_install_path=$(grep "MefEdgeSoftwareDir" /usr/lib/systemd/system/mef-edge-main.service | awk -F= '{print $2}');then
        logger_error "Get edge install path failed"
        return 1
    fi

    local edge_collect_shell_path=${edge_install_path}/run.sh
    if ! check_link "${edge_collect_shell_path}";then
        logger_error "MEF collect log shell path is invalid"
        return 1
    fi

    if [[ $("${OS_CMD_STAT}" -c %U "${edge_collect_shell_path}") != "root" ]]; then
       logger_error "edge collect shell path user must be root."
       return 1
    fi

    # 校验路径是否包含other的写权限
    local other_write_mod
    other_write_mod=$("${OS_CMD_STAT}" -c %A "${edge_collect_shell_path}" | "${OS_CMD_AWK}" '{print substr($1,9,1)}')
    if [[ "${other_write_mod}" != '-' ]]; then
        logger_error "other write mod is illegal."
        return 1
    fi

    if ! bash "${edge_collect_shell_path}" collectlog -module="${module}" -log_pack_path="${LOG_PATH}"/mef_edge.tar.gz;then
        logger_error "Collect MEF ${module} logs failed"
        return 1
    fi

    logger_info "Collect MEF ${module} logs Successfully"
    return 0
}

#collect  OS&Diver
function OS_log()
{
    if [ ! -d "/var/plog" ]; then
        logger_warn "can not find plog in var"
    fi

    if [ ! -d "/home/log/plog" ]; then
        logger_warn "can not find plog in home"
    fi

    if [ ! -d "/var/log" ]; then
        logger_warn "can not find log in var"
    fi

    if [ ! -d "/home/log/slog" ]; then
        logger_warn "can not find slog in home"
    fi

    tar --ignore-failed-read -czPf "${LOG_PATH}"/OS_Drivers.tar.gz \
    /var/log/audit.log \
    /var/log/messages \
    /var/log/secure \
    /var/plog/devm_scripts_run.log \
    /var/plog/ntp_service.log \
    /var/plog/ies_FAULT.log \
    /var/plog/upgrade.log \
    /var/plog/programs_write_alarm.log \
    /var/plog/ies_devm_drv.log \
    /var/plog/ies_devm_adapter.log \
    /var/plog/ies_lteswitch.log \
    /var/plog/programs_write_info.log \
    /home/log/slog \
    /home/log/plog/devm* \
    /home/log/plog/ntp* \
    /home/log/plog/ies* \
    /home/log/plog/upgrade* \
    /home/log/plog/programs* \
    /home/log/kbox_last_logs
    logger_info "Collect OS logs Successfully"
    return 0
}

#collect iBMA-Edge log
function MindXOM_log()
{
    if [ ! -d "/home/log/plog/ibma_edge" ]; then
        logger_warn "can not find ibma_edge in home"
    fi

    if [ ! -d "/var/plog/ibma_edge" ]; then
        logger_warn "can not find ibma_edge in var"
    fi

    if [ ! -d "/home/log/plog/redfish" ]; then
        logger_warn "can not find redfish in home"
    fi

    if [ ! -d "/var/plog/redfish" ]; then
        logger_warn "can not find redfish in var"
    fi

    if [ ! -d "/home/log/plog/manager" ]; then
        logger_warn "can not find manager in home"
    fi

    if [ ! -d "/var/plog/manager" ]; then
        logger_warn "can not find manager in var"
    fi

    if [ ! -d "/home/log/plog/web_edge" ]; then
        logger_warn "can not find web_edge in home"
    fi

    if [ ! -d "/var/plog/web_edge" ]; then
        logger_warn "can not find web_edge in var"
    fi

    tar --ignore-failed-read -czPf "${LOG_PATH}"/MindXOM.tar.gz \
    /var/plog/ibma_edge \
    /home/log/plog/ibma_edge \
    /var/plog/redfish \
    /home/log/plog/redfish \
    /var/plog/manager \
    /home/log/plog/manager \
    /var/plog/web_edge \
    /home/log/plog/web_edge
    logger_info "Collect MindXOM logs Successfully"
    return 0
}
