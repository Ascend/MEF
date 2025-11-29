#!/bin/bash

#***********************************************
#  Description: 文件或文件夹是否为软链接
#    Parameter: 1、文件或文件夹名
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
check_soft_link()
{
    if [[ $# -ne 1 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local input_name="$1"
    local format_name
    local real_name
    local normal_path="${input_name}"

    if [ "${input_name: -1}" == "/" ]
    then
        normal_path="${input_name:0:-1}"
    fi

    format_name=$(realpath -ms "${input_name}")
    real_name=$(readlink -m "${input_name}")
    if [[ "${format_name}" != "${real_name}" ]] || [[ "${real_name}" != "${normal_path}" ]]; then
        logger_error "File is soft link or illegal characters in path!"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 文件夹是否不包含特殊字符和相对路径
#    Parameter: 1、文件夹名
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
check_path()
{
    if [[ $# -ne 1 ]]; then
        logger_error "parameter error."
        return 1
    fi

    echo "$1" | grep -q -E '^[0-9a-zA-Z/-_]*$'
    local ret=$?
    return ${ret}
}

#***********************************************
#  Description: 文件或文件夹权限是否满足要求
#    Parameter: 1、文件或文件夹名
#               2、需要满足的权限，如：640
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
check_permission()
{
    if [[ $# -ne 2 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local file_name="$1"
    local input_permission="$2"
    local real_permission

    real_permission=$(stat -c %a "${file_name}")
    if [[ "${input_permission}" != "${real_permission}" ]]; then
        logger_error "Permission are not equal!"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 校验文件或文件夹属组是否满足要求
#    Parameter: 1、文件或文件夹名
#               2、需要满足的文件属组，如：0
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
check_group_owner()
{
    if [[ $# -ne 2 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local file_name="$1"
    local input_owner="$2"
    local real_owner

    real_owner=$(stat -c %g "${file_name}")
    if [[ "${input_owner}" != "${real_owner}" ]]; then
        logger_error "Group ID of owner are not equal!"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 校验文件或文件夹属主是否满足要求
#    Parameter: 1、文件或文件夹名
#               2、需要满足的文件属主，如：0
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
check_user_owner()
{
    if [[ $# -ne 2 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local file_name="$1"
    local input_user="$2"
    local real_user

    real_user=$(stat -c %u "${file_name}")
    if [[ "${input_user}" != "${real_user}" ]]; then
        logger_error "User ID of owner are not equal!"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 校验文件或文件夹属主是否满足要求
#    Parameter: 1、文件或文件夹名
#               2、需要满足的文件属主，如：0
#               3、需要满足的文件属组，如：0
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
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

#***********************************************
#  Description: 校验文件或文件夹属主是否满足要求
#    Parameter: 1、文件或文件夹名
#               2、需要满足的文件属主，如：0
#               3、需要满足的文件属组，如：0
#               4、需要满足的文件权限，如：750
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
check_owner_and_permission()
{
    if [[ $# -ne 4 ]]; then
        logger_error "parameter error."
        return 1
    fi

    local file_name="$1"
    local input_user="$2"
    local input_owner="$3"
    local input_permission="$4"
    local real_user
    local real_owner
    local real_permission

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

    real_permission=$(stat -c %a "${file_name}")
    if [[ "${input_permission}" != "${real_permission}" ]]; then
        logger_error "Permission are not equal!"
        return 1
    fi

    return 0
}

#***********************************************
#  Description: 校验IPV4地址是否满足要求
#    Parameter: 1、IPV4地址
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
function check_ipv4_is_valid() 
{
    local ip="$1"
    local ret="1"
    if [[ "${ip}" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
        # 按.分割，转成数组，方便下面的判断
        ip_array=()
        IFS=" " read -r -a ip_array <<< "${ip//\./ }"
        [[ "${ip_array[0]}" -le 255 && "${ip_array[1]}" -le 255 && "${ip_array[2]}" -le 255 && "${ip_array[3]}" -le 255 ]]
        ret=$?
    fi
    return ${ret}
}
