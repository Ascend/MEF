#!/bin/bash
CUR_DIR=$(dirname "$(readlink -f "$0")")
source /home/data/config/mindx_om_env.conf
source "${CUR_DIR}"/log_print.sh
source "${CUR_DIR}"/common_os_adapter.sh

init_log_path "/var/plog/manager" "manager_run.log"


#***********************************************
#  Description: 将OM的系统文件拷贝到指定分区，若有备区，则挂载并拷贝到备区，否则拷贝到当前系统分区
#  Parameter:
#       Input: $1: 挂载设备，$2: 挂载点
#       Output: NA
# Return Value:
#       0 -- 挂载成功
#       1 -- 挂载失败
#***********************************************
function mount_partition()
{
    if [[ ! -d "$2" ]]; then
        mkdir -p "$2"
    fi

    # 如果已经挂载，直接返回
    if df | grep -w "$1" >> /dev/null 2>&1; then
        return 0
    fi

    if ! mount -t ext4 "$1" "$2" >> /dev/null 2>&1; then
        logger_error "mount $1 to $2 failed"
        return 1
    fi
    return 0
}

#*****************************************************************************
# Description  : 挂载备区，将OM系统文件拷贝到备区
# Parameter:
#   input:  NA
#   output: NA
# Return Value : 0主区, 1备区1
#*****************************************************************************
function main()
{
    # OMSDK生效不拷贝到备区
    local upgrade_type="$1"
    local copy_type="$2"
    local om_src_path="$3"
    if [[ "${upgrade_type}" == "MindXOM" ]]; then
        # 拷贝OM的service、scripts、配置等系统文件到系统目录
        (
            export LD_LIBRARY_PATH="${om_src_path}"/lib:"${LD_LIBRARY_PATH}"
            export PYTHONPATH="${om_src_path}"/software/ibma:"${om_src_path}"/software/ibma/opensource/python
            python3 -u "${om_src_path}"/scripts/python/copy_om_sys_file.py "/" "${copy_type}" "${om_src_path}"
        )
        local ret=$?
        if [[ "${ret}" -ne 0 ]]; then
            logger_error "copy MindXOM sys file failed"
            return 1
        fi
        return 0
    fi

    # 获取uboot启动备区并挂载
    local back_dev_part_num=""
    back_dev_part_num=$(get_sys_back_dev_partition_num)
    local cur_uboot_back_area=""
    cur_uboot_back_area=$(get_sys_back_dev)
    if ! mount_partition  "${cur_uboot_back_area}" "/mnt/p${back_dev_part_num}"; then
        return 1
    fi

    # 拷贝OM的service、scripts、配置等系统文件到系统目录
    (
        export LD_LIBRARY_PATH="${om_src_path}"/lib:"${LD_LIBRARY_PATH}"
        export PYTHONPATH="${om_src_path}"/software/ibma:"${om_src_path}"/software/ibma/opensource/python
        python3 -u "${om_src_path}"/scripts/python/copy_om_sys_file.py "/mnt/p${back_dev_part_num}" "${copy_type}" "${om_src_path}"
    )

    local ret=$?
    if [[ "${ret}" -ne 0 ]]; then
        logger_error "copy MindXOM sys file failed"
        umount "${cur_uboot_back_area}"
        return 1
    fi

    # 拷贝完之后，解除挂载的系统备区
    logger_info "copy MindXOM sys file success"
    umount "${cur_uboot_back_area}"
    return 0
}

main "$@"
exit "$?"