#!/bin/bash
#check result
# 0: no alarm
# 1: device error
# 2: partion lose
# 4: device lose
# 8: mount failed
#16: mount error

check_soft_link()
{
    if [[ $# -ne 1 ]]; then
        return 1
    fi

    local input_name="$1"
    local format_name
    local real_name

    format_name=$(realpath -ms "${input_name}")
    real_name=$(readlink -m "${input_name}")
    if [[ "${format_name}" != "${real_name}" ]]; then
        return 1
    fi

    return 0
}

check_user_and_group_owner()
{
    if [[ $# -ne 3 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local file_name="$1"
    local input_user="$2"
    local input_owner="$3"
    local real_user
    local real_owner

    real_user=$(stat -c %u "${file_name}")
    if [[ "${input_user}" != "${real_user}" ]]; then
        logger_error "User ID of owner are not equal!"
        return 1
    fi

    real_owner=$(stat -c %g "${file_name}")
    if [[ "${input_owner}" != "${real_owner}" ]]; then
        logger_error "Group ID of owner are not equal!"
        return 1
    fi

    return 0
}

_log()
{
    LEVEL="$1"
    shift 1
    check_soft_link "${logfile}"
    local ret=$?
    if [[ -f "${logfile}" && "${ret}" -eq 0 ]]; then
        cur_date=$(date +"%Y-%m-%d %H:%M:%S")
        echo "${cur_date} [${LEVEL}]$1" >> "${logfile}"
    else
        cur_date=$(date +"%Y-%m-%d %H:%M:%S")
        >&2  echo "${cur_date} [${LEVEL}]$1"
    fi
}

logger_info()
{
    _log INFO "$@"
}

ROOT_USER_ID="0"
ROOT_GROUP_ID="0"
logfile="/var/plog/ibma_edge/om_scripts_run.log"

check_res=0
cfg_mount_file_site='/home/data/ies/mountCnf_site.ini'

if [[ ! -f "${cfg_mount_file_site}" ]]; then
    check_res=0
    echo "${check_res}"
    exit 0
fi

if ! check_soft_link "${cfg_mount_file_site}"; then
    logger_error "mountCnf_site.ini contain soft links!"
    exit 1
fi

dev_list=()
IFS=" " read -r -a dev_list <<< "$(< "${cfg_mount_file_site}" grep "dev" | tr -d ' ' | sort | uniq | awk -F '=' '{print $1,$2}' | tr '\n' ' ')"
list_num=${#dev_list[@]}
for ((i=0; i<list_num; i=i+2))
do
    if [[ -e ${dev_list[$i]} ]]; then
        if [[ "${dev_list[$i]}" =~ "disk" ]]; then
            site_part=${dev_list[$i]##*/}
            disk_dev_recover_cmd="/usr/local/scripts/disk_dev_recover.sh"
            if ! check_user_and_group_owner "${disk_dev_recover_cmd}" ${ROOT_USER_ID} ${ROOT_GROUP_ID}; then
                logger_error "the disk_dev_recover.sh owner is not right."
                exit 1
            fi

            if ! check_soft_link "${disk_dev_recover_cmd}"; then
                logger_error "disk_dev_recover.sh contain soft links!"
                exit 1
            fi

            raw_part=$("${disk_dev_recover_cmd}" "${site_part}")
            if [[ "$raw_part" != "" ]]; then
                dev_list[$i]="/dev/$raw_part"
            fi
        fi

        res1=$(mount -l | grep "${dev_list[$i]}")
        if [ "${res1}" != "" ] && [ $((i + 1)) -lt "${list_num}" ]; then
            j=$((i+1))
            if ! cd "${dev_list[$j]}" > /dev/NULL;then
                check_res=$((check_res|16))
                logger_info "[0x10]:${dev_list[$j]} not exist!"
            else
                dst_dir=$(pwd)
                res=$(mount -l | grep "${dev_list[$i]}" | grep "${dst_dir}")
                if [ "$res" == "" ];then
                    check_res=$((check_res|16))
                    logger_info "[0x10]:${dev_list[$i]}"
                fi
            fi
            cd - > /dev/NULL || exit
        else
            #no mount point exist,
            check_res=$((check_res|8))
            logger_info "[0x8]:${dev_list[$i]}"
        fi
    else
        #device or partion exist
        res1=$(echo "${dev_list[$i]}" | grep mmcblk)
        if [ "$res1" != "" ];then
            #this is a sd card or in emmc flash
            if [ -e "${dev_list[$i]:0:12}" ];then
                #partion not exist but device exist!
                check_res=$((check_res|2))
                logger_info "[0x2]:${dev_list[$i]}"
            else
                #device not exist!
                check_res=$((check_res|4))
                logger_info "[0x4]:${dev_list[$i]}"
            fi
            continue
        fi

        res2=$(echo "${dev_list[$i]}" | grep sd)
        if [ "$res2" != "" ];then
            #this is a usb or hdd
            if [ -e "${dev_list[$i]:0:8}" ];then
                #partion not exist but device exist!
                check_res=$((check_res|2))
                logger_info "[0x2]:${dev_list[$i]}"
            else
                #device not exist!
                check_res=$((check_res|4))
                logger_info "[0x4]:${dev_list[$i]}"
            fi
            continue
        fi

        res3=$(echo "${dev_list[$i]}" | grep disk)
        if [ "$res3" != "" ];then
            #this is a usb or hdd
            if [ -h "${dev_list[$i]:0:11}" ];then
                #partion not exist but device exist!
                check_res=$((check_res|2))
                logger_info "[0x2]:${dev_list[$i]}"
            else
                #device not exist!
                check_res=$((check_res|4))
                logger_info "[0x4]:${dev_list[$i]}"
            fi
            continue
        fi

        check_res=$((check_res|1))
        logger_info "[0x1]:${dev_list[$i]}"
    fi
done
echo $check_res
