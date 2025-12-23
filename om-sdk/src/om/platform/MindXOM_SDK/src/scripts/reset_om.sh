#!/bin/bash
# 说明：此脚本是恢复出厂设置时使用，被驱动调用和触发，且不能失败，否则设备有变砖风险。
GOLDEN_DEV="/dev/mmcblk0p1"
GOLDEN_MOUNT_DIR="/mnt/p1"
SYS_MAIN_DEV="/dev/mmcblk0p2"
SYS_MAIN_MOUNT_DIR="/mnt/p2"
SYS_BACK_DEV="/dev/mmcblk0p3"
SYS_BACK_MOUNT_DIR="/mnt/p3"
HOME_DATA_DEV="/dev/mmcblk0p4"
HOME_DATA_MOUNT_DIR="/home/data"
MINDX_HOME_DEV="/dev/mmcblk0p6"
MINDX_HOME_DIR="/usr/local/mindx"
OM_RESET_LOG="om_restore_msg.log"
WHITEBOX_BACK_DIR="/home/data/WhiteboxConfig"
P1_WHITEBOX_DIR="${GOLDEN_MOUNT_DIR}/whiteboxConfig"
OM_RESET_DIR="${MINDX_HOME_DIR}/MindXOMReset"
RESET_OM_SERVICE="reset-om.service"
RESET_OM_FLAG="om_reset_flag"
RESET_OM_MINIOS_FLAG="om_reset_minios_flag"
# 操作类型，默认恢复出厂
MODE_TYPE="reset_factory"
# 恢复出厂设置
readonly MODE_RESET_FACTORY="reset_factory"
# 最小系统恢复
readonly MODE_RESET_MINIOS="reset_minios"


#***********************************************
#  功能描述: 挂载分区
#  参数: $1: 挂载设备，$2: 挂载点
# 返回值: 0表示挂载成功， 1表示挂载失败
#***********************************************
function mount_partition()
{
    if [[ ! -d "$2" ]]; then
        mkdir -p "$2"
    fi

    mount -t ext4 "$1" "$2" > /dev/null 2>&1;
    return 0
}

#***********************************************
#  Description: 拷贝OM系统文件到p2、p3分区
#  Parameter:
#       Input: $1：拷贝的目标目录
#       Output: NA
# Return Value:
#       0 -- 成功
#       1 -- 失败
#***********************************************
function copy_om_init_files()
{
    # 拷贝om-init的service、对应脚本到系统主备区
    local dest_file="${SYS_MAIN_MOUNT_DIR}/usr/lib/systemd/system/${RESET_OM_SERVICE}"
    if [ -L "${dest_file}" ]; then
        unlink "${dest_file}"
    fi
    cp -f "${OM_RESET_DIR}/software/service_main/${RESET_OM_SERVICE}" "${dest_file}"
    ln -sf "/usr/lib/systemd/system/${RESET_OM_SERVICE}" "${SYS_MAIN_MOUNT_DIR}/etc/systemd/system/multi-user.target.wants/${RESET_OM_SERVICE}"

    dest_file="${SYS_BACK_MOUNT_DIR}/usr/lib/systemd/system/${RESET_OM_SERVICE}"
    if [ -L "${dest_file}" ]; then
        unlink "${dest_file}"
    fi
    cp -f "${OM_RESET_DIR}/software/service_main/${RESET_OM_SERVICE}" "${dest_file}"
    ln -sf "/usr/lib/systemd/system/${RESET_OM_SERVICE}" "${SYS_BACK_MOUNT_DIR}/etc/systemd/system/multi-user.target.wants/${RESET_OM_SERVICE}"

}

#***********************************************
#  Description: 将OM软件包拷贝解压至临时目录，进行恢复出厂
#  Parameter:
#       Input: NA
#       Output: NA
# Return Value:
#       0 -- 成功
#       1 -- 失败
#***********************************************
function unpack_om_package()
{
    # 解压OM软件包
    local om_package=""
    rm -rf "${OM_RESET_DIR}"
    mkdir -p "${OM_RESET_DIR}"
    if [[ "${MODE_TYPE}" == "${MODE_RESET_MINIOS}" ]]; then
        om_package=$(find "${GOLDEN_MOUNT_DIR}/firmware" -maxdepth 1 -name "Ascend-om_*_linux-aarch64.tar.gz")
    else
        om_package=$(find "${GOLDEN_MOUNT_DIR}" -maxdepth 1 -name "Ascend-om_*_linux-aarch64.tar.gz")
    fi

    if [[ ! -f "${om_package}" ]]; then
        echo "Can not find om package"
        return 1
    fi

    if ! tar --no-same-owner -zxf "${om_package}" -C "${OM_RESET_DIR}"; then
        echo "unpack ${om_package} to ${OM_RESET_DIR} failed"
        return 1
    fi

    # 如果存在integrate_om_files.sh, 说明om包代码结构拆成了A500-A2-om.tar.gz和om-sdk.tar.gz
    if [[ -f "${OM_RESET_DIR}/scripts/integrate_om_files.sh" ]];then
        bash "${OM_RESET_DIR}/scripts/integrate_om_files.sh" "${OM_RESET_DIR}" "${OM_RESET_DIR}/om-sdk.tar.gz" \
        "${OM_RESET_DIR}/A500-A2-om.tar.gz" "${OM_RESET_DIR}/version.xml"
    fi

    return 0
}

#***********************************************
#  Description: 拷贝白牌配置
#  Parameter:
#       Input: NA
#       Output: NA
# Return Value:
#       0 -- 成功
#       1 -- 失败
#***********************************************
function copy_om_whitebox_config()
{
    # 没有白牌配置，就不白牌化
    if [[ ! -d  "${P1_WHITEBOX_DIR}" ]]; then
        echo "no need to whitebox processing"
        return 0
    fi

    if ! rm -rf "${WHITEBOX_BACK_DIR}" >& /dev/null; then
        echo "clear whitebox config dir ${WHITEBOX_BACK_DIR} failed, please check"
    fi

    if ! mkdir -p "${WHITEBOX_BACK_DIR}" >& /dev/null; then
        echo "create ${WHITEBOX_BACK_DIR} failed, please check"
    fi

    if ! cp -rf "${P1_WHITEBOX_DIR}"/* "${WHITEBOX_BACK_DIR}/" >& /dev/null; then
        echo "save ${WHITEBOX_BACK_DIR} failed, please check"
    fi

    return 0
}

#*******************************************************
#  Description: 从Golden分区拷贝PKI证书到Nginx配置目录
#  Parameter:
#       Input: NA
#       Output: NA
# Return Value:
#       0 -- 成功
#       1 -- 失败
#*******************************************************
function copy_pki_cert_from_golden_dev()
{
    local pki_certs_path="${GOLDEN_MOUNT_DIR}/om_certs"

    if [[ ! -d  "${pki_certs_path}" ]]; then
        echo "pki certs not exists"
        return 0
    fi

    local nginx_default_path="${HOME_DATA_MOUNT_DIR}/config/default"
    if [[ ! -d "${nginx_default_path}" ]]; then
        mkdir -p "${nginx_default_path}"
        chmod 700 "${nginx_default_path}"
    fi

    local file_list=(
        "om_cert.keystore"
        "om_cert_backup.keystore"
        "server_kmc.cert"
        "server_kmc.priv"
        "server_kmc.psd"
        "om_alg.json"
    )
    for file in "${file_list[@]}"; do
        local src_file="${pki_certs_path}/${file}"
        local dest_file="${nginx_default_path}/${file}"
        if [ -L "${dest_file}" ]; then
            unlink "${dest_file}"
        fi
        if ! cp -f "${src_file}" "${dest_file}"; then
            echo "copy pki certs config to nginx default dir failed"
            return 1
        fi
        chmod 600 "${dest_file}"
        chown -h nobody:nobody "${dest_file}"
    done

    return 0
}

#***********************************************
#  Description: 挂载OM需要的分区
#  Parameter:
#       Input: NA
#       Output: NA
# Return Value:
#       0 -- 成功
#       1 -- 失败
#***********************************************
function mount_om_partitions()
{
    # 挂载golden分区、系统主区、系统备区、配置区、日志区
    if ! mount_partition "${GOLDEN_DEV}" "${GOLDEN_MOUNT_DIR}"; then
        echo "mount ${GOLDEN_DEV} to ${GOLDEN_MOUNT_DIR} failed"
        umount "${GOLDEN_DEV}"
        return 1
    fi

    if ! mount_partition "${SYS_MAIN_DEV}" "${SYS_MAIN_MOUNT_DIR}"; then
        echo "mount ${SYS_MAIN_DEV} to ${SYS_MAIN_MOUNT_DIR} failed"
        umount "${SYS_MAIN_DEV}"
        return 1
    fi

    if ! mount_partition "${SYS_BACK_DEV}" "${SYS_BACK_MOUNT_DIR}"; then
        echo "mount ${SYS_BACK_DEV} to ${SYS_BACK_MOUNT_DIR} failed"
        umount "${SYS_BACK_DEV}"
        return 1
    fi

    if ! mount_partition "${HOME_DATA_DEV}" "${HOME_DATA_MOUNT_DIR}"; then
        echo "mount ${HOME_DATA_DEV} to ${HOME_DATA_MOUNT_DIR} failed"
        umount "${HOME_DATA_DEV}"
        return 1
    fi

    if ! mount_partition "${MINDX_HOME_DEV}" "${MINDX_HOME_DIR}"; then
        echo "mount ${MINDX_HOME_DEV} to ${MINDX_HOME_DIR} failed"
        umount "${MINDX_HOME_DEV}"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 解除挂载OM需要的分区
#  Parameter:
#       Input: NA
#       Output: NA
# Return Value:
#       0 -- 成功
#       1 -- 失败
#***********************************************
function umount_om_partitions()
{
    # 解载golden分区、系统主区、系统备区、日志区
    umount "${GOLDEN_DEV}"
    umount "${SYS_MAIN_DEV}"
    umount "${SYS_BACK_DEV}"
    umount "${HOME_DATA_DEV}"
    umount "${MINDX_HOME_DEV}"

    return 0
}

#***********************************************
#  Description: 入参校验
#    Parameter:
#        Input: 系统盘符字符串
#       Output:
# Return Value:
#       0 -- 合法
#       1 -- 非法
#***********************************************
function check_param_valid()
{
    if [[ $# != 2 ]]; then
        echo "The number of parameters should be equal to 2, current is $#."
        return 1
    fi

    local sys_dev_path="$1"
    if ! [[ "${sys_dev_path}" =~ ^/dev/[a-z0-9]{1,255}$ ]];then
        echo "The dev path is invalid: $1 not match /dev/[a-z0-9]{1,255}"
        return 1
    fi

    if [[ ! -b "${sys_dev_path}" ]]; then
        echo "The dev path ${sys_dev_path} is not block file or not exists"
        return 1
    fi

    local mode_type="$2"
    if [[ "${mode_type}" != "${MODE_RESET_FACTORY}" && "${mode_type}" != "${MODE_RESET_MINIOS}" ]]; then
        echo "The reset mode can only be ${MODE_RESET_FACTORY} or ${MODE_RESET_MINIOS}, current is ${mode_type}"
        return 1
    fi

    if [[ "${mode_type}" == "${MODE_RESET_MINIOS}" ]] && ! [[ "${sys_dev_path}" =~ ^/dev/mmcblk0$ ]]; then
        # 最小系统恢复功能只支持emmc启动场景
        echo "The reset mode ${mode_type} is not supported, because current boot dev is not /dev/mmcblk0"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 初始化系统盘符路径
#    Parameter:
#        Input: 系统盘符字符串
#       Output:
# Return Value:
#       0 -- 成功
#       1 -- 失败
#***********************************************
function init_sys_dev_path()
{
    local sys_dev_path="$1"
    local mode_type="$2"

    if [[ "${sys_dev_path}" =~ ^/dev/mmcblk0$ ]]; then
        echo "The current boot dev is emmc, add suffix with char p"
        sys_dev_path="${sys_dev_path}p"
    fi

    if [[ "${mode_type}" == "${MODE_RESET_MINIOS}" ]]; then
        # 最小系统恢复场景的Golden分区为p7分区
        GOLDEN_DEV="${sys_dev_path}7"
        GOLDEN_MOUNT_DIR="/home/package"
        MODE_TYPE="${MODE_RESET_MINIOS}"
    else
        GOLDEN_DEV="${sys_dev_path}1"
        MODE_TYPE="${MODE_RESET_FACTORY}"
    fi
    if [[ ! -b "${GOLDEN_DEV}" ]]; then
        echo "The boot dev path ${GOLDEN_DEV} is not block file or not exists"
        return 1
    fi
    echo "Setting GOLDEN_DEV to ${GOLDEN_DEV} success"
    echo "Setting GOLDEN_MOUNT_DIR to ${GOLDEN_MOUNT_DIR} success"
    echo "Setting MODE_TYPE to ${MODE_TYPE} success"

    SYS_MAIN_DEV="${sys_dev_path}2"
    if [[ ! -b "${SYS_MAIN_DEV}" ]]; then
        echo "The sys main dev path ${SYS_MAIN_DEV} is not block file or not exists"
        return 1
    fi
    echo "Setting SYS_MAIN_DEV to ${SYS_MAIN_DEV} success"

    SYS_BACK_DEV="${sys_dev_path}3"
    if [[ ! -b "${SYS_BACK_DEV}" ]]; then
        echo "The sys back dev path ${SYS_BACK_DEV} is not block file or not exists"
        return 1
    fi
    echo "Setting SYS_BACK_DEV to ${SYS_BACK_DEV} success"

    HOME_DATA_DEV="${sys_dev_path}4"
    if [[ ! -b "${HOME_DATA_DEV}" ]]; then
        echo "The home data dev path ${HOME_DATA_DEV} is not block file or not exists"
        return 1
    fi
    echo "Setting HOME_DATA_DEV to ${HOME_DATA_DEV} success"

    MINDX_HOME_DEV="${sys_dev_path}6"
    if [[ ! -b "${MINDX_HOME_DEV}" ]]; then
        echo "The mindx dev path ${MINDX_HOME_DEV} is not block file or not exists"
        return 1
    fi
    echo "Setting MINDX_HOME_DEV to ${MINDX_HOME_DEV} success"

    return 0
}

#***********************************************
#  Description: OM恢复出厂设置流程
#    Parameter:
#        Input: NA
#       Output: NA
# Return Value: 0
#      Caution: NA
#***********************************************
function main()
{
    echo -e "\n****************************************************"
    echo "start to reset MindXOM"

    if ! check_param_valid "$@"; then
        echo "Parameters invalid: $*"
        return 1
    fi

    if ! init_sys_dev_path "$@"; then
        echo "Init sys dev path failed"
        return 1
    fi

    # 挂载golden分区、系统主区、系统备区、日志区
    if ! mount_om_partitions; then
        echo "mount om partitions failed"
        return 1
    fi

    # 从golden分区拷贝OM软件包
    if ! unpack_om_package; then
        echo "unpack om package failed"
        return 1
    fi

    # 从golden分区拷贝白牌配置
    if ! copy_om_whitebox_config; then
        echo "om whitebox processing failed"
        return 1
    fi

    # 从golden分区拷贝pki证书到nginx工作目录，失败不影响整个恢复出厂流程
    copy_pki_cert_from_golden_dev

    # 拷贝OM系统文件到p2、p3分区
    if ! copy_om_init_files; then
        echo "copy om sys files failed"
        return 1
    fi

    # 将恢复出厂日志拷贝出去
    if [ -L "${HOME_DATA_MOUNT_DIR}/${OM_RESET_LOG}" ]; then
        unlink "${HOME_DATA_MOUNT_DIR}/${OM_RESET_LOG}"
    fi
    cp "${GOLDEN_MOUNT_DIR}/${OM_RESET_LOG}" "${HOME_DATA_MOUNT_DIR}/${OM_RESET_LOG}"

    # 创建恢复出厂标记
    touch "${HOME_DATA_MOUNT_DIR}/${RESET_OM_FLAG}"

    if [[ "${MODE_TYPE}" == "${MODE_RESET_MINIOS}" ]]; then
        # 创建恢复最小系统标记
        touch "${HOME_DATA_MOUNT_DIR}/${RESET_OM_MINIOS_FLAG}"
    fi

    # 解挂OM所需要的分区
    umount_om_partitions

    echo "end to reset MindXOM"
    echo -e "\n****************************************************"

    return 0
}

main "$@"
exit $?