#!/bin/bash

# get current path
CURRENT_PATH=$(cd "$(dirname "$0")" || exit 3; pwd)
# Load log module
source "${CURRENT_PATH}"/log_print.sh
# Load comm_checker module
source "${CURRENT_PATH}"/comm_checker.sh

ntp_conf_file="/run/ntp.conf"
ntp_etc_file="/etc/ntp.conf"
ntpd_file="/etc/sysconfig/ntpd"
logfile="/var/plog/ntp_service.log"
ntp_enable_file="/home/data/ies/NTPEnable.ini"
ifcfg_dir="/etc/sysconfig/network-scripts"
ubuntu_ifcfg_dir="/run/systemd/network"
OS_NAME=$(< "/etc/os-release" grep "^NAME=" | awk -F "=" '{print $2}' | tr -d '"')

init_log_path "/var/plog" "ntp_service.log"

## 1、输入参数错误  2、命令或配置文件不存在   3、获取到的端口IP错误

#*****************************************************************************
# Prototype    : list_local_IP
# Description  : 呈现已经配置过的IP列表
# Parameter:
#   input:       NA
# Return Value : NA
#  History        :
#  1.Date         : 2019/8/16
#    Modification : Created function
#*****************************************************************************
function list_local_IP()
{
    local IP_num=0
    local IP_temp=""

	if [[ "${OS_NAME}" = "EulerOS" ]]; then
		for file in "${ifcfg_dir}"/ifcfg-eth*; do
            IP_temp=$(< "${file}" grep "^IPADDR" | awk -F "=" '{print $2}' | tr -d '"')
            logger_info "IP_temp is: ${IP_temp}"
            # 解决恢复出厂设置场景下,未保留网卡IP为空的情况
            if ! check_IP_addr "${IP_temp}"; then
                continue
            fi
            local_IP_list["${IP_num}"]="${IP_temp}"
            ((IP_num++))
        done
    elif [[ "${OS_NAME}" = "openEuler" ]]; then
    	local open_euler_ip_array=()
		for file in "${ifcfg_dir}"/ifcfg-eth*; do
			IFS=" " read -r -a open_euler_ip_array <<< "$(< "${file}" grep "^IPADDR" | awk -F "=" '{print $2}' | tr -d '"')"
			for ipv4 in "${open_euler_ip_array[@]}"; do
				logger_info "Get sysconfig file [${file}] ip is: ${ipv4}"
				# 解决恢复出厂设置场景下,未保留网卡IP为空的情况
				if ! check_IP_addr "${ipv4}"; then
					continue
				fi
				local_IP_list["${IP_num}"]="${ipv4}"
				((IP_num++))
			done
        done
    elif [[ "${OS_NAME}" = "Ubuntu" ]]; then
        for file in "${ubuntu_ifcfg_dir}"/10-netplan*; do
            IP_temp=$(< "${file}" grep "Address" | awk -F "=" '{print $2}' | tr -d '/')
            logger_info "IP_temp is: ${IP_temp}"
            # 解决恢复出厂设置场景下,未保留网卡IP为空的情况
            if ! check_IP_addr "${IP_temp}"; then
                continue
            fi
            local_IP_list["${IP_num}"]="${IP_temp}"
            ((IP_num++))
        done
    else
		logger_error "The OS [${OS_NAME}] is not supported."
    	return 1
    fi

    logger_info "list_local_IP success. IP num is: ${IP_num}"
    return 0
}

#*****************************************************************************
# Prototype    : update_conf_file
# Description  : 根据输入参数修改配置文件 /etc/ntp.conf
# Parameter:
#   input:   NTP_remote_servers NTP_local_servers
#   output: NA
# Return Value : NA
#  History        :
#  1.Date         : 2019/5/9
#    Modification : Created function
#*****************************************************************************
function update_conf_file()
{
    local num=1
    local ip_num=0

    cp -a "${ntp_etc_file}" "${ntp_conf_file}"

    sed -i 's/^#.*//g' "${ntp_conf_file}"        #删除注释行
    sed -i '/^ *$/d' "${ntp_conf_file}"           #删除空行
    sed -i '/restrict/d' "${ntp_conf_file}"

    while read -r line
    do
        if [[ "${line}" =~ "server" ]] && [[ "${delete_server_ip_flag}" -eq 0 ]] || [[ "${line}" =~ "tos" ]] || \
            [[ "${line}" =~ "includefile" ]] || [[ "${line}" =~ "key" ]]; then  #不是特定关键字
            ((num++))
        else
            sed -i "${num} d" "${ntp_conf_file}"    #删除本行
        fi
    done <${ntp_conf_file}

    sed -i '1adriftfile /var/lib/ntp/drift' "${ntp_conf_file}"

    if [[ "${ntp_client_enable}" -eq 1 ]];then
        # 判断server_ip是否存在
        sed -i 's/prefer//g' "${ntp_conf_file}"  # 先将prefer替换

        ip_get=$(< "${ntp_conf_file}" grep "${NTP_remote_servers}")
        if [ "$ip_get" != "" ];then
            sed -i "/${NTP_remote_servers}/d" "${ntp_conf_file}"
        fi

        echo "server ${NTP_remote_servers} prefer" >> "${ntp_conf_file}"
        echo "fudge ${NTP_remote_servers} stratum 10" >> "${ntp_conf_file}"

        if [[ x"$NTP_remote_servers_bak" != x"" ]]; then
            ip_get=$(< "${ntp_conf_file}" grep "${NTP_remote_servers_bak}")
            if [ "${ip_get}" != "" ]; then
                sed -i "/${NTP_remote_servers_bak}/d" "${ntp_conf_file}"
            fi
            echo "server ${NTP_remote_servers_bak} " >> "${ntp_conf_file}"
            echo "fudge ${NTP_remote_servers_bak} stratum 10" >> "${ntp_conf_file}"
        fi

        echo "restrict ${NTP_remote_servers} kod nomodify notrap nopeer noquery" >> "${ntp_conf_file}"
        if [[ x"${NTP_remote_servers_bak}" != x"" ]];then
            echo "restrict $NTP_remote_servers_bak kod nomodify notrap nopeer noquery" >> "${ntp_conf_file}"
        fi
        list_local_IP

        # 开放本地IP
        for ((ip_num=0; ip_num<${#local_IP_list[@]}; ip_num++));
        do
            echo "restrict ${local_IP_list[ip_num]}" >> "${ntp_conf_file}"
        done
    fi

    # 首先限制所有接入的IP
    {
        echo "restrict -4 default kod nomodify notrap nopeer noquery"
        echo "restrict -6 default kod nomodify notrap nopeer noquery"

        #安全要求关闭monlist命令
        echo "disable monitor"
        #安全要求关闭ntp服务器对操作码2的响应
        echo "restrict -4 default noquery ignore"

        # 对本机没有任何限制
        echo "restrict 127.0.0.1"
        echo "restrict -6 ::1"
        echo "logfile ${logfile}"
    } >> "${ntp_conf_file}"

    # 解决全零监听
    sed -i '/^interface.*$/d' ${ntp_conf_file}
    echo "interface ignore wildcard" >> ${ntp_conf_file}

    if [[ -n "${NTP_local_servers}" ]]; then
        NTPD_OPTION=$(< "${ntpd_file}" grep "OPTIONS" | awk -F "=" '{ print $2 }' | tr -d '"')
        if [[ "${NTPD_OPTION}" == "-g" ]]; then
            sed -i 's/-g/-4 -g/g' "${ntpd_file}"
        fi
        echo "interface listen $NTP_local_servers" >> "${ntp_conf_file}"
    else
        # local ip 为空，默认随机选取IP监听
        for ((ip_num=0; ip_num<${#local_IP_list[@]}; ip_num++));
        do
            echo "interface listen ${local_IP_list[ip_num]}" >> "${ntp_conf_file}"
            break
        done
    fi

    cp -a "${ntp_conf_file}" "${ntp_etc_file}"
    rm -f "${ntp_conf_file}"

    return 0
}

#*****************************************************************************
# Prototype    : start_ntp_service
# Description  ：启动NTP服务
# Parameter:
#   input:   NTP_remote_servers NTP_local_servers
#   output: NA
# Return Value : NA
#  History        :
#  1.Date         : 2019/5/9
#    Modification : Created function
#*****************************************************************************
function start_ntp_service()
{
    # 首先修改配置文件
    logger_info "ntp.conf $(date +"%Y-%m-%d %H:%M:%S") modify"
    logger_info "ntp.conf NTP_remote_servers: ${NTP_remote_servers}; NTP_remote_servers_bak: ${NTP_remote_servers_bak}; NTP_local_servers: ${NTP_local_servers}"

    update_conf_file "$1" "$2"

    # 启动NTP 服务
    if [[ "${OS_NAME}" = "EulerOS" || "${OS_NAME}" = "openEuler" ]];then
        systemctl start ntpd >/dev/null 2>&1
    elif [[ "${OS_NAME}" = "Ubuntu" ]]; then
        systemctl start ntp >/dev/null 2>&1
    else
    	logger_error "The OS [${OS_NAME}] is not supported."
    	return 1
    fi

    return 0
}

#*****************************************************************************
# Prototype    : check_IP_addr
# Description  : IP地址合理性校验
# Parameter:
#   input:   IP address
#   output: NA
# Return Value : NA
#  History        :
#  1.Date         : 2019/5/9
#    Modification : Created function
#*****************************************************************************
function check_IP_addr()
{
    #IP地址必须为全数字且为三位数
    if ! echo "$1" |grep "^[0-9]\{1,3\}\.\([0-9]\{1,3\}\.\)\{2\}[0-9]\{1,3\}$" > /dev/null; then
        logger_error "input id address format wrong"
        return 1
    fi

    ipaddr="$1"
    a=$(echo "${ipaddr}" | awk -F . '{print $1}')   #以"."分隔，取出每个列的值
    b=$(echo "${ipaddr}" | awk -F . '{print $2}')
    c=$(echo "${ipaddr}" | awk -F . '{print $3}')
    d=$(echo "${ipaddr}" | awk -F . '{print $4}')
    for num in "$a" "$b" "$c" "$d"
    do
        if [ "${num}" -gt 255 ] || [ "${num}" -lt 0 ]     #每个数值必须在0-255之间
        then
            logger_error "IP addr input is wrong"
            return 1
        fi
    done
    return 0
}

#*****************************************************************************
# Prototype    : read_enable_flag
# Description  : 首先获取启动标识
# Parameter:
#   input:   NA
#   output: NA
# Return Value : NA
#  History        :
#  1.Date         : 2019/8/15
#    Modification : Created function
#*****************************************************************************
function read_enable_flag()
{
    local client_enable_tmp

    # 首先获取启动标识
    if [ ! -f "$ntp_enable_file" ];then
        logger_error "ntp service enabled file not exist"
        exit 2
    else
        ntp_server_enable=0
        client_enable_tmp=$(< "${ntp_enable_file}" grep "'NTPClientEnable': True")
        if [[ -n "${client_enable_tmp}" ]]; then
            ntp_client_enable=1
        fi
        logger_info "ntp service enable flag, ntp_server_enable: $ntp_server_enable, ntp_client_enable: $ntp_client_enable"
    fi

    if [[ "${ntp_client_enable}" -eq 0 ]];then
        logger_error "please check $ntp_enable_file, disabled ntp_client_enable: $ntp_client_enable"
        exit 1
    fi
}

#*****************************************************************************
# Prototype    : check_local_server
# Description  : 检查输入的local IP是否已经配置到系统中
# Parameter:
#   input:       NTP_local_servers
# Return Value : NA
#  History        :
#  1.Date         : 2019/8/16
#    Modification : Created function
#*****************************************************************************
function check_local_server()
{
    local IP_temp=""
    local local_ip=$1

	if [[ "${OS_NAME}" = "EulerOS" ]]; then
		 for file in "${ifcfg_dir}"/ifcfg-eth*; do
            IP_temp=$(< "${file}" grep "^IPADDR" | awk -F "=" '{print $2}' | tr -d '"')
            if [[ "${IP_temp}" == "${local_ip}" ]];then
                logger_info "Check local server ip [${local_ip}] is in ${file}."
                return 0
            fi
        done
	elif [[ "${OS_NAME}" = "openEuler" ]]; then
		local open_euler_ip_array=()
		for file in "${ifcfg_dir}"/ifcfg-eth*; do
            IFS=" " read -r -a open_euler_ip_array <<< "$(< "${file}" grep "^IPADDR" | awk -F "=" '{print $2}' | tr -d '"')"
            for item in "${open_euler_ip_array[@]}"; do
            	if [[ "${item}" == "${local_ip}" ]];then
                	logger_info "Check local server ip [${local_ip}] is in ${file}."
                	return 0
            	fi
            done
        done
    elif [[ "${OS_NAME}" = "Ubuntu" ]]; then
        local ubuntu_ip_array=()
        for file in "${ubuntu_ifcfg_dir}"/10-netplan*; do
            IFS=" " read -r -a ubuntu_ip_array <<< "$(< "${file}" grep "Address=" | awk -F "=" '{print $2}' | awk -F "/" '{print $1}' | tr '\n' ' ')"
            for temp_ip in "${ubuntu_ip_array[@]}"; do
                if [[ "${temp_ip}" == "${local_ip}" ]];then
                    logger_info "Check local server ip [${local_ip}] is in ${file}."
                    return 0
                fi
            done
        done
    else
    	logger_error "The OS [${OS_NAME}] is not supported."
    	return 1
    fi

	logger_error "Check local server ip failed. input local ip [${local_ip}] can not find in system."
	return 1
}

#*****************************************************************************
# Prototype    : run
# Description  : 检查输入参数是否正确
# Parameter:
#   input:   升级选项
#   output: NA
# Return Value : NA
#  History        :
#  1.Date         : 2019/3/8
#    Modification : Created function
#*****************************************************************************
function run()
{
    if [ $# -eq 0 ];then
        logger_error "Check input paramter failed."
        return 1
    fi

    # 根据参数1执行不同的分支
    case $1 in
    start)
        # 获取使能的启动参数
        read_enable_flag

        NTP_remote_servers=$(echo "$2" | awk -F ';' '{print $1}')
        NTP_remote_servers_bak=$(echo "$2" | awk -F ';' '{print $2}')

        if ! check_IP_addr "$NTP_remote_servers"; then
            logger_error "check_IP_addr failed. input server ip is $NTP_remote_servers. please input right format IP."
            return 1
        fi

        if [ x"$NTP_remote_servers_bak" != x"" ]; then
            if ! check_IP_addr "$NTP_remote_servers_bak"; then
                logger_error "check_IP_addr failed. input server ip is $NTP_remote_servers_bak. please input right format IP."
                return 1
            fi
        fi

       # 判断$3是否为“-f”，强制删除原来的IP
        if [[ "$3" != "-f" ]];then
            NTP_local_servers="$3"
            if [[ -n "${NTP_local_servers}" ]]; then
                if ! check_IP_addr "${NTP_local_servers}"; then
                    logger_info "check_IP_addr failed. input server ip is ${NTP_local_servers}."
                    return 1
                fi
                # 判断IP是否在系统中已经配置
                if ! check_local_server "${NTP_local_servers}"; then
                    logger_error "local servers can not find ${NTP_local_servers}."
                    return 1
                fi
            fi
        fi

        # 判断是否有参数为“-f”，强制删除原来的IP
        if [[ "$4" == "-f" ]] || [[ "$3" == "-f" ]];then
            delete_server_ip_flag=1
        fi

        start_ntp_service "$NTP_remote_servers" "$NTP_local_servers"
        ;;
    stop)
    	if [[ "${OS_NAME}" = "EulerOS" || "${OS_NAME}" = "openEuler" ]];then
        	systemctl stop ntpd >/dev/null 2>&1
    	elif [[ "${OS_NAME}" = "Ubuntu" ]]; then
        	systemctl stop ntp >/dev/null 2>&1
    	else
    		logger_error "The OS [${OS_NAME}] is not supported."
    		return 1
    	fi
        sed -i '/^interface.*$/d' ${ntp_etc_file}
        ;;
    --ntp_usage) ntpd --help
        ;;
    esac

 return 0
}

#*****************************************************************************
# Prototype    : check_ntp_file
# Description  : 检查NTP使用的文件是否存在软链接
# Parameter:
#   input:  NA
#   output: NA
# Return Value : 0 or 1
#*****************************************************************************
function check_ntp_file()
{
    local nfs_file=(
        "${ntp_conf_file}"
        "${ntp_etc_file}"
        "${ntpd_file}"
        "${logfile}"
        "${ntp_enable_file}"
    )

    for path in "${nfs_file[@]}"; do
        if ! check_soft_link "${path}"; then
            logger_error "${path} contain soft links!"
            return 1
        fi
    done

    return 0
}

#######################################################
#main
#######################################################

NTP_local_servers=""              # 用哪个IP向外同步时间。如果不指定，所有能连接到小站上的设备均能同步时间
NTP_remote_servers=""             # 从哪里获取时间
NTP_remote_servers_bak=""

local_IP_list=()
delete_server_ip_flag=0

# server、client启动标识，0：不启动，1：启动。全局变量
ntp_server_enable=0
ntp_client_enable=0

# 先判断配置文件和命令是否存在
if [ ! -f "$ntp_etc_file" ];then
    logger_error "ntp.conf not exist"
    exit 2
else
    if ! ntpd --help > /dev/null; then
        logger_error "ntpd command not exist"
        exit 2
    fi
fi

if ! check_ntp_file; then
    logger_error "check ntp file fail."
    exit 2
fi

# 输入参数检查并执行
run "$@"
ret=$?
if [ "${ret}" -ne 0 ];then
    logger_error "run failed. ret=$ret"
    exit "${ret}"
fi

exit 0
