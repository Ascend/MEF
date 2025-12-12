#!/bin/bash

# set process umask right
umask 0077

UPGRADE_FINISHED_FLAG="/home/data/upgrade_finished_flag"
OM_BACKUP_DIR="/usr/local/mindx/MindXOM_bak"
IES_DIR_PATH="/home/data/ies"
SYNC_FLAG="/home/data/om_sync_flag"
ASCEND_DMI_PATH="/usr/local/Ascend/toolbox"
CANN_NNRT_PATH="/usr/local/Ascend/cann"
LOG_SYNC_PATH="/home/log/om_sync"

# get current path
CURRENT_PATH=$(
    cd "$(dirname "$0")" || exit 3
    pwd
)

#***********************************************
#  Description: 检查命令结果，记录日志，非0直接退出
#    Parameter:
#        Input: 命令的返回码、命令描述
#       Output: NA
# Return Value: 0
#      Cattion: NA
#***********************************************
function check_result()
{
    local ret="$1"
    local message="$2"

    if [ "$ret" -eq 0 ]; then
       logger_info "$message success."
       return 0
    else
       logger_error "$message failed."
       exit 1
    fi
}

#***********************************************
#  Description: 设置高危配置文件
#    Parameter:
#        Input: NA
#       Output: NA
# Return Value: 0
#      Cattion: NA
#***********************************************
function set_ability_policy()
{
    if [ ! -f "${UPGRADE_FINISHED_FLAG}" ]; then
        return 0
    fi

    logger_info "start set ability policy file"
    local config_dir='/home/data/config'
    local config_file="${config_dir}/ability_policy.json"
    if [ ! -d "${config_dir}" ]; then
        mkdir -p "${config_dir}"
        check_result $? "make ability policy config dir"
        chmod 755 "${config_dir}"
        check_result $? "chmod ability policy config dir"
        logger_info "create ability policy config dir success"
    fi

    local om_dir="${OM_UPGRADE_DIR}"
    local default_allow_all_config_file="$om_dir/software/ibma/config/ability_policy_allow_all.json"
    if [ -f "${OM_RESET_FLAG}" ]; then
        cp -f "${default_allow_all_config_file}" "${config_file}"
        check_result $? "get rest flag: copy disable all ability policy config to config_file"
        chmod 644 "${config_file}"
        check_result $? "chmod ability policy config_file"
    else
        if [ ! -f "${config_file}" ]; then
            cp -f "${default_allow_all_config_file}" "${config_file}"
            check_result $? "get upgrade flag: copy allow all ability policy config to config_file"
            chmod 644 "${config_file}"
            check_result $? "chmod ability policy config_file"
        else
            logger_info "get upgrade flag: ability policy config_file already exist"
        fi
    fi
    logger_info "set ability policy file success"

    return 0
}

function set_file_immutable_attr()
{
    if [[ -d "${OM_WORK_DIR}" ]]; then
        find "${OM_WORK_DIR}"/ -name "*.sh" -exec chattr +i {} \;
        find "${OM_WORK_DIR}"/ -name "*.py" -exec chattr +i {} \;
        find "${OM_WORK_DIR}"/ -name "*.so*" -type f -exec chattr +i {} \;
        find "${OM_WORK_DIR}"/ -name "*.xml" -exec chattr +i {} \;
        logger_info "set file immutable attr success"
    else
        logger_info "${OM_WORK_DIR} not exist, skip setting file immutable attr"
    fi
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
            "os_cmd_euleros2.0.conf"
            "os_cmd_ubuntu22.04.conf"
            "os_cmd_openEuler_22.03.conf"
        )

    # 1、 筛选出源目录下的os_cmd*.conf文件
    for file in "${os_cmd_file_list[@]}"; do
        local file_path="${src_dirpath}/${file}"
        os_name=$(< "${file_path}" grep "OS_NAME" | awk -F "=" '{print $2}' | tr -d '"')
        os_version_id=$(< "${file_path}" grep "OS_VERSION_ID" | awk -F "=" '{print $2}' | tr -d '"')
        logger_info "os_name is: ${os_name}  os_version_id is: ${os_version_id}"

        # 2、与系统信息进行匹配校验
        if ! check_os_info "${os_name}" "${os_version_id}"; then
            continue
        fi
        logger_info "os_name is: ${os_name}  os_version_id is: ${os_version_id}"

        # 3、调用拷贝函数进行拷贝
        cp -f "${file_path}" "${dst_filepath}"
        chmod 644 "${dst_filepath}"
        logger_info "replace os cmd conf from ${file_path} to ${dst_filepath} success"
        return 0
    done

    # 4、如果配置文件匹配失败，则默认将os_cmd_euleros2.0.conf文件内容作为默认的配置
    logger_error "replace os cmd conf to ${dst_filepath} failed,will replace default"
    cp -f "${src_dirpath}"/os_cmd_euleros2.0.conf "${dst_filepath}"
    chmod 644 "${dst_filepath}"

    return 0
}

function update_system_config()
{
    # 拷贝环境变量
    source "${OM_WORK_DIR}/scripts/safe_common.sh"

    local home_data_config="/home/data/config"
    cp -f "${OM_WORK_DIR}/config/mindx_om_env.conf" "${home_data_config}/mindx_om_env.conf"
    chmod 640 "${home_data_config}/mindx_om_env.conf"
    safe_replace_os_cmd_conf "${OM_WORK_DIR}/config" "${home_data_config}/os_cmd.conf"

    # 重新刷新为交换之后的配置文件
    source "${home_data_config}/os_cmd.conf"
    source "${home_data_config}/mindx_om_env.conf"
}

function upgrade_om()
{
    # 交换工作区和备区
    if [ -f "${UPGRADE_FINISHED_FLAG}" ]; then
        logger_info "UPGRADE_FINISHED_FLAG exists, will upgrade MindXOM!"
        set_ability_policy

        # 交换work目录与upgrade目录
        if [[ -d "${OM_WORK_DIR}" ]]; then
            # 升级情况下，执行数据库迁移，由开发者保证版本间迁移成功
            bash "${OM_UPGRADE_DIR}/scripts/migrate.sh" "migrate"

            mv "${OM_WORK_DIR}" "${OM_BACKUP_DIR}"
            mv "${OM_UPGRADE_DIR}" "${OM_WORK_DIR}"
            mv "${OM_BACKUP_DIR}" "${OM_UPGRADE_DIR}"

            # 调用mef生效脚本
            local effect_py="${OM_WORK_DIR}/scripts/python/effect_mef.py"
            (
                export LD_LIBRARY_PATH="${OM_WORK_DIR}"/lib:"${LD_LIBRARY_PATH}"
                export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python:"${OM_WORK_DIR}"/scripts/python
                python3 "${effect_py}"
            )
        else
            mv "${OM_UPGRADE_DIR}" "${OM_WORK_DIR}"
        fi

        # 更新系统配置
        update_system_config

        rm "${UPGRADE_FINISHED_FLAG}"
        chattr -i -R "${OM_UPGRADE_DIR}"
        rm "${OM_UPGRADE_DIR}" -rf

        set_file_immutable_attr

        unzip_firmware

        logger_info "MindXOM upgrade done!"
    fi
}

function check_om_env_path()
{
    local om_env_paths=(
        "${OM_WORK_DIR}"
        "${OM_BACKUP_DIR}"
        "${OM_UPGRADE_DIR}"
        "${IES_DIR_PATH}"
        "/usr/lib64/"
        "/home/data/config"
    )

    for path in "${om_env_paths[@]}"; do
        if ! check_soft_link "${path}"; then
            logger_error "${path} contain soft links!"
            return 1
        fi

        # 不存在则跳过属主校验
        if [[ ! -e "${path}" ]]; then
            continue
        fi
        local root_user_id="0"
        local root_group_id="0"
        if ! check_user_and_group_owner "${path}" "${root_user_id}" "${root_group_id}"; then
            logger_error "the ${path} owner is not right!"
            return 1
        fi

        # 校验路径是否包含other的写权限
        local other_write_mod
        other_write_mod=$(${OS_CMD_STAT} -c %A "${path}" | ${OS_CMD_AWK} '{print substr($1,9,1)}')
        if [[ "${other_write_mod}" != '-' ]]; then
            logger_error "other write mod is illegal."
            return 1
        fi
    done
}

function init_default_log_path()
{
    # ibma-edge
    local ibma_edge_dir="/var/plog/ibma_edge"
    init_log_path "${ibma_edge_dir}" "om_scripts_run.log"
    local sync_ibma_dir="${LOG_SYNC_PATH}/ibma_edge"
    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2'; then
        local ibma_logs=("om_scripts_run.log" "om_platform_run.log" "seceye_agent_run.log")
        for (( i= 0; i < "${#ibma_logs[@]}"; i++)); do
            unlink "${ibma_edge_dir}/${ibma_logs[i]}"
            mv "${sync_ibma_dir}/${ibma_logs[i]}" "${ibma_edge_dir}/${ibma_logs[i]}"
        done
    fi

    # upgrade.log
    local upgrade_log="/var/plog/upgrade.log"
    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2'; then
        unlink "${upgrade_log}"
        mv "${LOG_SYNC_PATH}/upgrade.log" "${upgrade_log}"
    fi

    # manager
    local manager_dir="/var/plog/manager"
    local sync_manager_dir="${LOG_SYNC_PATH}/manager"
    mkdir -p "${manager_dir}" && chmod 700 "${manager_dir}"
    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2'; then
        local manager_logs=("manager_run.log" "manager_operate.log")
        for (( i= 0; i < "${#manager_logs[@]}"; i++)); do
            unlink "${manager_dir}/${manager_logs[i]}"
            mv "${sync_manager_dir}/${manager_logs[i]}" "${manager_dir}/${manager_logs[i]}"
        done
    fi

    # 创建redfish日志目录
    local redfish_log_dir="/var/plog/redfish"
    local sync_redfish_dir="${LOG_SYNC_PATH}/redfish"
    if ! check_soft_link "${redfish_log_dir}"; then
        logger_error "${redfish_log_dir} contain soft links!"
        return 1
    fi

    if [[ ! -d "${redfish_log_dir}" ]]; then
        mkdir -p "${redfish_log_dir}"
        chmod 700 "${redfish_log_dir}"
    fi
    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2'; then
        local redfish_logs=("redfish_run.log" "redfish_operate.log" "redfish_platform_run.log")
        for (( i= 0; i < "${#redfish_logs[@]}"; i++)); do
            # 先删除再移动，防止软链接提权
            unlink "${redfish_log_dir}/${redfish_logs[i]}"
            mv "${sync_redfish_dir}/${redfish_logs[i]}" "${redfish_log_dir}/${redfish_logs[i]}"
        done
    fi
    chown -hR MindXOM:MindXOM "${redfish_log_dir}"

    # ntp_service日志初始化:解决系统生成时默认为644权限问题
    local ntp_server_log="/var/plog/ntp_service.log"
    if [[ -L "${ntp_server_log}" ]]; then
        unlink "${ntp_server_log}"
    fi

    if [[ ! -f "${ntp_server_log}" ]]; then
        touch "${ntp_server_log}"
    fi

    chmod 640 "${ntp_server_log}"
    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2'; then
        mv "${LOG_SYNC_PATH}/ntp_service.log" "${ntp_server_log}"
    fi

    # nginx目录
    local home_web_edge_dir="/var/plog/web_edge"
    local sync_web_edg_dir="${LOG_SYNC_PATH}/web_edge"
    if ! check_soft_link "${home_web_edge_dir}"; then
        logger_error "${home_web_edge_dir} contain soft links!"
        return 1
    fi

    if [[ ! -d "${home_web_edge_dir}" ]]; then
        mkdir -p "${home_web_edge_dir}"
        chmod 700 "${home_web_edge_dir}"
    fi

    if npu-smi info -t product -i 0 | grep 'Atlas 500 A2'; then
        local web_edge_logs=("access.log" "error.log" "web_edge_run.log")
        for (( i= 0; i < "${#web_edge_logs[@]}"; i++)); do
            # 先删除再移动，防止软链接提权
            unlink "${home_web_edge_dir}/${web_edge_logs[i]}"
            mv "${sync_web_edg_dir}/${web_edge_logs[i]}" "${home_web_edge_dir}/${web_edge_logs[i]}"
        done
    fi
    chown -hR "$(id "nobody" -u)":"$(id "nobody" -g)" "${home_web_edge_dir}"

    return 0
}

function check_root_env_path()
{
    if [[ $# -ne 1 ]]; then
        echo "parameter error."
        return 1
    fi

    local path="$1"
    local format_name
    local real_name
    local user_id
    local group_id

    # 校验软链接
    format_name=$(${OS_CMD_REALPATH} -ms "${path}")
    real_name=$(${OS_CMD_READLINK} -m "${path}")
    if [[ "${format_name}" != "${real_name}" ]]; then
        echo "path is soft link!"
        return 2
    fi

    # 校验属主
    if [[ ! -e "${path}" ]]; then
        return 3
    fi

    user_id=$(${OS_CMD_STAT} -c %u "${path}")
    group_id=$(${OS_CMD_STAT} -c %g "${path}")
    if [[ -z "${user_id}" ]] || [[ -z "${group_id}" ]]; then
        echo "user or group not exist"
        return 5
    fi

    if [[ "${user_id}" != "0" ]] && [[ "${group_id}" != "0" ]]; then
        echo "The path owner check failed. path=${path}"
        return 6
    fi

    # 校验路径是否包含other的写权限
    local other_write_mod
    other_write_mod=$(${OS_CMD_STAT} -c %A "${path}" | ${OS_CMD_AWK} '{print substr($1,9,1)}')
    if [[ "${other_write_mod}" != '-' ]]; then
        echo "other write mod is illegal."
        return 7
    fi

    return 0
}

function check_init_root_path()
{
    local os_cmd_path="/home/data/config/os_cmd.conf"
    local om_env_cfg_path="/home/data/config/mindx_om_env.conf"
    if [ ! -f "${os_cmd_path}" ] || [ ! -f "${om_env_cfg_path}" ]; then
        echo "${os_cmd_path} or ${om_env_cfg_path} not exist or is symbolic link!"
        exit 1
    fi
    source "${os_cmd_path}"
    source "${om_env_cfg_path}"

    local root_env_paths=(
            "${CURRENT_PATH}"
            "/var/plog"
            "/usr/lib64"
            "/home/data"
        )
    local check_ret=""
    for path in "${root_env_paths[@]}"; do
        check_root_env_path "${path}"
        check_ret=$?
        if [[ "${check_ret}" -ne 0 ]]; then
            echo "${path} check_root_env_path failed! ${check_ret}"
            exit "${check_ret}"
        fi
    done
}

function unzip_firmware()
{
    local p7_path="/home/package/firmware"
    local firmware_path="${p7_path}/firmware.zip"

    if [ -f "${firmware_path}" ]; then
        find "${p7_path}" -name "*.tar.gz" -type f -exec rm -f {} \;
        find "${p7_path}" -name "vercfg.xml*" -type f -exec rm -f {} \;
        unzip -o -d "${p7_path}" "${firmware_path}"
        rm "${firmware_path}"
    fi
}

function sync_files()
{
    # 主备同步
    if [ -f "${SYNC_FLAG}" ]; then
        # 复制一份到临时目录，用来做主备同步
        local tmp_toolbox_path="/run/upgrade/toolbox"
        rm -rf "${tmp_toolbox_path}"
        mkdir -p "${tmp_toolbox_path}"
        cp -r "${ASCEND_DMI_PATH}" "$(dirname "${tmp_toolbox_path}")"

        local tmp_cann_path="/run/upgrade/cann"
        rm -rf "${tmp_cann_path}"
        mkdir -p "${tmp_cann_path}"
        cp -r "${CANN_NNRT_PATH}" "$(dirname "${tmp_cann_path}")"

        bash "${OM_WORK_DIR}/scripts/copy_om_sys_file.sh" "Ascend-firmware" "backup_area" "${OM_WORK_DIR}"
        rm "${SYNC_FLAG}"
    fi
}

main()
{
    # Load log module
    source "${CURRENT_PATH}"/log_print.sh
    # Load comm_checker module
    source "${CURRENT_PATH}"/comm_checker.sh

    check_init_root_path

    if [[ -f "${OM_WORK_DIR}/lib/liblpeblock.so" ]];then
        export LD_PRELOAD="${OM_WORK_DIR}/lib/liblpeblock.so"
    fi

    # 日志初始化
    init_default_log_path

    if ! check_om_env_path; then
        logger_error "check MindXOM env path fail."
        return 1
    fi

    upgrade_om

    # 主备同步
    sync_files

    return 0
}

main
