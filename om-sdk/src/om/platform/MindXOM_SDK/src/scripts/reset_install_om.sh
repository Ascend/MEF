#!/bin/bash
HOME_DATA_DIR="/home/data"
OM_RESET_LOG="${HOME_DATA_DIR}/om_restore_msg.log"
RESET_OM_FLAG="${HOME_DATA_DIR}/om_reset_flag"
RESET_OM_MINIOS_FLAG="${HOME_DATA_DIR}/om_reset_minios_flag"
OM_RESET_DIR="/usr/local/mindx/MindXOMReset"
OM_WORK_DIR="/usr/local/mindx/MindXOM"
OM_RESET_INSTALL_SH="${OM_RESET_DIR}/install.sh"
OM_SYNC_SHELL="${OM_WORK_DIR}/scripts/copy_om_sys_file.sh"
RESET_OP_LOG_DIR="/var/plog/manager"
RESET_OP_LOG_FILE="${RESET_OP_LOG_DIR}/manager_operate.log"
RESET_OM_SERVICE="reset-om.service"
WHITEBOX_BACK_DIR="/home/data/WhiteboxConfig"
ASCEND_DMI_PATH="/usr/local/Ascend/toolbox"
CANN_NNRT_PATH="/usr/local/Ascend/cann"
P1_MOUNT_PATH="/mnt/p1"
P1_DEV_PATH="/dev/mmcblk0p1"
# 软件包的存储盘符路径: reset_factory-p1分区; reset_minios-p7
PKG_DEV_PATH="${P1_DEV_PATH}"
PKG_MOUNT_PATH="${P1_MOUNT_PATH}"

source "${OM_RESET_DIR}"/scripts/common_os_adapter.sh

function check_is_reset_minios()
{
    if [[ -f "${RESET_OM_MINIOS_FLAG}" ]]; then
        return 0
    fi
    return 1
}

function check_is_reset_factory()
{
    if ! check_is_reset_minios; then
        return 0
    fi
    return 1
}

function install_ascend_dmi()
{
    if ! check_boot_from_emmc; then
        echo "current boot from m2, no need install Ascend-Dmi, skip it"
        return 0
    fi

    local sys_dev_path=""
    sys_dev_path=$(get_sys_dev_path_suffix)

    if check_is_reset_minios; then
        echo "The mode is reset_minios"
        PKG_DEV_PATH="${sys_dev_path}7"
        PKG_MOUNT_PATH="/home/package"
    else
        echo "The mode is reset_factory"
        PKG_DEV_PATH="${sys_dev_path}1"
        PKG_MOUNT_PATH="${P1_MOUNT_PATH}"
    fi

    if [[ ! -b "${PKG_DEV_PATH}" ]]; then
        echo "The partition dev path not exist: ${PKG_DEV_PATH}"
        return 1
    fi

    if [[ ! -d "${PKG_MOUNT_PATH}" ]]; then
        mkdir -p "${PKG_MOUNT_PATH}"
    fi

    echo "Setting PKG_DEV_PATH to ${PKG_DEV_PATH} success"
    echo "Setting PKG_MOUNT_PATH to ${PKG_MOUNT_PATH} success"
    if check_is_reset_factory; then
        echo "Current mode is reset_factory, need mount partition"
        mount -t ext4 "${PKG_DEV_PATH}" "${PKG_MOUNT_PATH}" > /dev/null 2>&1;
    else
        echo "Current mode is reset_minios, no need mount partition"
    fi

    # 解压安装toolbox、cann软件包
    local toolbox_package=""
    rm -rf "${ASCEND_DMI_PATH}"
    mkdir -p "${ASCEND_DMI_PATH}"

    local cann_package=""
    rm -rf "${CANN_NNRT_PATH}"
    mkdir -p "${CANN_NNRT_PATH}"

    local package_dir_path="${PKG_MOUNT_PATH}"
    if check_is_reset_minios; then
        package_dir_path="${PKG_MOUNT_PATH}/firmware"
        echo "Current mode is reset_minios, package_dir_path is ${package_dir_path}"
    else
        echo "Current mode is reset_factory, package_dir_path is ${package_dir_path}"
    fi

    toolbox_package=$(find "${package_dir_path}" -maxdepth 1 -name "Ascend-mindx-toolbox_*.tar.gz")
    cann_package=$(find "${package_dir_path}" -maxdepth 1 -name "Ascend-cann-nnrt_*.tar.gz")
    if [[ ! -f "${toolbox_package}" ]]; then
        echo "Can not find toolbox package: ${toolbox_package}"
        return 1
    fi
    if [[ ! -f "${cann_package}" ]]; then
        echo "Can not find cann package: ${cann_package}"
        return 1
    fi

    tar --no-same-owner -zxf "${toolbox_package}" -C "${ASCEND_DMI_PATH}" > /dev/null 2>&1;
    echo "source ${ASCEND_DMI_PATH}/set_env.sh" >> "/etc/profile"

    # 解压cann
    tar --no-same-owner -zxf "${cann_package}" -C "${CANN_NNRT_PATH}" > /dev/null 2>&1;

    # 复制一份到临时目录，用来做主备同步
    local tmp_toolbox_path="/run/upgrade/toolbox"
    rm -rf "${tmp_toolbox_path}"
    mkdir -p "${tmp_toolbox_path}"
    cp -r "${ASCEND_DMI_PATH}" "$(dirname "${tmp_toolbox_path}")"

    local tmp_cann_path="/run/upgrade/cann"
    rm -rf "${tmp_cann_path}"
    mkdir -p "${tmp_cann_path}"
    cp -r "${CANN_NNRT_PATH}" "$(dirname "${tmp_cann_path}")"

    if check_is_reset_factory; then
        umount "${PKG_MOUNT_PATH}"
        echo "Current mode is reset_factory, umount operation success: ${PKG_MOUNT_PATH}"
    else
        echo "Current mode is reset_minios, no need umount operation"
    fi
}

main()
{
    # 如果是恢复出厂脚本触发，需要安装OM
    if [[ -f "${RESET_OM_FLAG}" ]]; then
        if [[ ! -d "${RESET_OP_LOG_DIR}" ]]; then
            mkdir "${RESET_OP_LOG_DIR}" -p
        fi
        chmod 700 "${RESET_OP_LOG_DIR}"

        if [ -L "${RESET_OP_LOG_FILE}" ]; then
            unlink "${RESET_OP_LOG_FILE}"
        fi

        if [ ! -f "${RESET_OP_LOG_FILE}" ]; then
            touch "${RESET_OP_LOG_FILE}"
        fi
        chmod 640 "${RESET_OP_LOG_FILE}"

        # 给白牌配置目录添加i属性
        if [[ -d "${WHITEBOX_BACK_DIR}" ]]; then
            chattr +i -R "${WHITEBOX_BACK_DIR}" >& /dev/null
        fi

        # 安装OM
        bash "${OM_RESET_INSTALL_SH}"

        # 记录恢复出厂日志
        if [[ -f "${OM_RESET_LOG}" ]]; then
            cat "${OM_RESET_LOG}" >> "${RESET_OP_LOG_FILE}"
        else
            local local_ip=""
            local_ip=$(/usr/bin/who am i | cut -d \( -f 2 | cut -d \) -f 1)
            if [[ -z "${local_ip}" ]]; then
                local_ip="LOCAL"
            fi

            local local_name=""
            local_name=$(/usr/bin/whoami)
            if [[ -z "${local_name}" ]]; then
                local_name="root"
            fi

            local cur_date=""
            cur_date=$(date +"%Y-%m-%d %H:%M:%S,%3N")
            echo "[${cur_date}] [${local_name}@${local_ip}] $(echo "Restore defaults system success" | sed 's/\(\(\/\w\+\)\+\/\)/****\//g')" >> "${RESET_OP_LOG_FILE}"
        fi

        # 安装ascend-dmi
        install_ascend_dmi

        # 进行主备同步
        bash "${OM_SYNC_SHELL}" "Ascend-firmware" "backup_area" "${OM_WORK_DIR}"

        # 清楚恢复出厂的文件
        rm -rf "${OM_RESET_LOG}"
        rm -rf "${OM_RESET_DIR}"
        rm -rf "${RESET_OM_FLAG}"
        rm -rf "${RESET_OM_MINIOS_FLAG}"
        rm -rf "/usr/lib/systemd/system/${RESET_OM_SERVICE}"
        unlink "/etc/systemd/system/multi-user.target.wants/${RESET_OM_SERVICE}"
        rm -f "/home/package/firmware/reset_om.sh"
        rm -rf "/home/package/whiteboxConfig"
        rm -f "/home/package/om_restore_msg.log"
        rm -rf "/home/package/om_certs"

    fi
}
main