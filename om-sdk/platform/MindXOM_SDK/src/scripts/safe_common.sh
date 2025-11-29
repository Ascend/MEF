#!/bin/bash

function safe_change_mode() {

    if [ $# -lt 2 ] || [ $# -gt 3 ] || [[ $# -gt 2  && "$3" != "-R" ]]; then
        echo "safe chmod parameter error"
        return 1
    fi

    local mode="$1"
    local path="$2"
    local parameter_r
    if [[ $# -gt 2  && "$3" == "-R" ]]; then
        parameter_r="$3"
        if [[ -d "${path}" && "${UID}" == "0" ]]; then
            if [ "${path: -1}" == "/" ]; then
                path="${path:0:-1}"
            fi

            find "${path}" | while IFS= read -r file
            do
                if is_need_change_owner "${file}"; then
                    return 1
                fi
            done

            chmod "${mode}" "${path}" "${parameter_r}"
            return 0
        fi
    fi

    # path is exist and valid
    if [ ! -e "${path}" ] || [ "${path:0:1}" != "/" ]; then
        echo "path is not exist"
        return 1
    fi

    if [[ "${UID}" == "0" ]]; then
        if [ -L "${path}" ]; then
            echo "The path is a soft link!"
            return 1
        fi
        local cur_path="${path}"
        while [ "${cur_path}" != "/" ]; do
            if is_need_change_owner "${cur_path}"; then
                return 1
            fi
            cur_path=$(dirname "${cur_path}")
        done
        chmod "${mode}" "${path}" "${parameter_r}"
        return 0
    else
        chmod "${mode}" "${path}" "${parameter_r}"
        return 0
    fi
}


function check_black_dirs() {
    local rm_full_black_dirs="/  /etc  /run"
    local path="$1"
    # check black dirs
    for dir in ${rm_full_black_dirs}
    do
        if [ "${dir}" == "${path}" ]; then
            echo "The path is a black path!"
            return 1
        fi
    done
}

function safe_rm() {
    local path="$1"
    if [ ! -e "${path}" ] || [ "${path:0:1}" != "/" ]; then
      echo "$path not exist"
      return 1
    fi

    check_black_dirs "${path}"

    # check root/root group
    group_id=$(id -g)
    if [ "${UID}" != "0" ] && [ "${group_id}" != "0" ]; then
      rm -rf "${path}"
      return 0
    fi

    # link file
    if [ -L "${path}" ] && [ -f "${path}" ]; then
       unlink "${path}"
       return 0
    fi

    # link dir
    if [ -L "${path}" ] && [ -d "${path}" ] && [[ $(stat -c %U "${path}") != "root" ]]; then
        echo "${path} does not comply with security rules." && return 1
    fi

    # Loop
    local cur_path
    cur_path=$(dirname "${path}")
    while true; do
      if [ "${cur_path}" == '/' ]; then
        break
      fi
      if [ -L "${cur_path}" ] && [[ $(stat -c %U "${cur_path}") != "root" ]]; then
        echo "${cur_path} does not comply with security rules." && return 1
      fi
      cur_path=$(dirname "${cur_path}")
    done
    rm -rf "${path}"
}


function safe_cp() {
    if [ $# -lt 2 ]; then
        echo "safe cp parameter error"
        return 1
    fi
    local srcpath
    local dstpath

    srcpath=$(realpath -s "$1")
    dstpath=$(realpath -s "$2")

    # check root:root group
    group_id=$(id -g)
    if [ "${UID}" != "0" ] && [ "${group_id}" != "0" ]; then
        cp "$@"
        return 0
    fi

    # srcpath is link and owner is not root
    if [ -L "${srcpath}" ] && [[ $(stat -c %U "${srcpath}") != "root" ]]; then
        echo "${srcpath} does not comply with security rules." && return 1
    fi

    # Loop src
    local cur_src_path
    cur_src_path=$(dirname "${srcpath}")
    while true; do
        if [ "${cur_src_path}" == '/' ] || [ ! -e "${cur_src_path}" ]; then
            break
        fi

        if [ -L "${cur_src_path}" ] && [[ $(stat -c %U "${cur_src_path}") != "root" ]]; then
            echo "${cur_src_path} does not comply with security rules." && return 1
        fi
        cur_src_path=$(dirname "${cur_src_path}")
    done

    # Loop dest
    local cur_dest_path="${dstpath}"
    while true; do
        if [ "${cur_dest_path}" == '/' ] || [ ! -e "${cur_dest_path}" ]; then
            break
        fi
        if [ -L "${cur_dest_path}" ] && [[ $(stat -c %U "${cur_dest_path}") != "root" ]]; then
            echo "${cur_dest_path} does not comply with security rules." && return 1
        fi
        cur_dest_path=$(dirname "${cur_dest_path}")
    done

    cp "$@"
    return 0
}

function safe_mkdir() {
    local dir_path="$1"
    local multilevel="$2"
    if [ -z "${dir_path}" ]; then
        echo "the parameter is null."
        return 2
    fi
    if [ -d "${dir_path}" ]; then
        echo "${dir_path} already exists."
        return 2
    fi
    local last_path
    last_path=$(dirname "${dir_path}")
    while true; do
        if [ "${last_path}" == '/' ]; then
            break
        fi
        if [ "${multilevel}" == "no" ] && [ ! -d "${last_path}" ]; then
            echo "${last_path}  does not exist."
            return 2
        fi
        if [ "$(id -u)" == "0" ] || [ "$(id -g)" == "0" ]; then
            if [ -h "${last_path}" ] && [[ $(stat -c %U "${last_path}") != "root" ]]; then
                echo "${last_path} does not comply with security rules."
                return 2
            fi
        fi
        last_path=$(dirname "${last_path}")
    done
    mkdir -p "${dir_path}"
}

function safe_mv() {
    # check root/root group
    group_id=$(id -g)
    if [ "${UID}" != "0" ] && [ "${group_id}" != "0" ]; then
        mv "$@"
        return 0
    fi

    local src
    local dest

    src=$(realpath -s "$1")
    dest=$(realpath -s "$2")

    check_black_dirs "${src}"

    # link file
    if [ -L "${src}" ] && [ -f "${src}" ]; then
        mv "$@"
        return 0
    fi

    # link dir and owner is not root
    if [ -L "${src}" ] && [ -d "${src}" ] && [[ $(stat -c %U "${src}") != "root" ]]; then
        echo "${src} does not comply with security rules." && return 1
    fi

    # Loop src
    local cur_src_path
    cur_src_path=$(dirname "${src}")
    while true; do
        if [ "${cur_dest_path}" == '/' ] || [ ! -e "${cur_dest_path}" ]; then
            break
        fi

        if [ -L "${cur_src_path}" ] && [[ $(stat -c %U "${cur_src_path}") != "root" ]]; then
            echo "${cur_src_path} does not comply with security rules." && return 1
        fi
        cur_src_path=$(dirname "${cur_src_path}")
    done

    # Loop dest
    local cur_dest_path
    cur_dest_path=$(dirname "${dest}")
    while true; do
        if [ "${cur_dest_path}" == '/' ] || [ ! -e "${cur_dest_path}" ]; then
            break
        fi
        if [ -L "${cur_dest_path}" ] && [[ $(stat -c %U "${cur_dest_path}") != "root" ]]; then
            echo "${cur_dest_path} does not comply with security rules." && return 1
        fi
        cur_dest_path=$(dirname "${cur_dest_path}")
    done

    mv "$@"
}


# Determine whether to run the chown parameter: path
# If path is soft link and the owner is not root, the chown command is not executed and returns.
function is_need_change_owner()
{
    local path="$1"
    # the return formats of files and directories are different from command 'ls -l ',
    # but the return formats of the soft link directory and files are the same.
    # Therefore, the system determines that the directory is not a soft link and directly returns the file,
    if [ ! -L "${path}" ]; then
        return 1
    fi

    local user_id
    local group_id

    user_id=$(stat -c %u "${path}")
    group_id=$(stat -c %g "${path}")
    if [ -z "${user_id}" ] || [ -z "${group_id}" ]; then
        echo "user or group not exist"
        return 0
    fi

    if [ "${user_id}" != "0" ] && [ "${group_id}" != "0" ]; then
        echo "The path is a soft link and the owner is not root, not need chown"
        return 0
    fi

    return 1
}


function safe_change_owner()
{
    if [ $# -lt 3 ] || [ $# -gt 4 ] || [[ $# -gt 3  && "$4" != "-R" ]]; then
        echo "safe chown parameter error"
        return 1
    fi

    local user="$1"
    local group="$2"
    local path="$3"
    local parameter_r
    if [[ $# -gt 3  && "$4" == "-R" ]]; then
        parameter_r="$4"
    fi
    # path is exist and valid
    if [ ! -e "${path}" ] || [ "${path:0:1}" != "/" ]; then
        echo "path is not exist"
        return 1
    fi
    local user_id
    local group_id

    user_id=$(id "${user}" -u)
    group_id=$(id "${group}" -g)
    if [ -z "${user_id}" ] || [ -z "${group_id}" ]; then
        echo "user:${user} or group:${group} not exist"
        return 1
    fi
    # if Escalate the permission, Run the chown command without checking the soft link
    if [ "${user_id}" == "0" ] || [ "${group_id}" == "0" ]; then
        chown -h "${parameter_r}" "${user}:${group}" "${path}"
        return 0
    fi
    # Recursively determine the parent directory need chown.
    local cur_path="${path}"
    while [ "${cur_path}" != "/" ]; do
        if is_need_change_owner "${cur_path}"; then
            return 1
        fi
        cur_path=$(dirname "$cur_path")
    done

    if [[ -z "${parameter_r}" ]]; then
        chown -h "${user_id}:${group_id}" "${path}"
    else
        chown -h "${parameter_r}" "${user_id}:${group_id}" "${path}"
    fi

    return 0
}

#*****************************************************************************
# Prototype    : safe_check_tarfile
# Description  : 检查tar文件
# Parameter:
#   input:   tar文件
#   output: NA
# Return Value : NA
#  History        :
#  1.Date         : 2021/09/23
#    Modification : Created function
#*****************************************************************************
function safe_check_tarfile() {
    local file_path="$1"
    # 判断参数是否为空
    if [ -n "${file_path}" ]; then
        # 判断文件是否存在
        if [ ! -f "${file_path}" ]; then
            echo "the ${file_path} is not file or not exist."
            return 1
        fi

        if tar -tvf "${file_path}" | awk -F" " '{print $1}' | grep 's'; then
            echo "the ${file_path} include SUID permissions."
            return 1
        fi
        return 0
    else
        echo "${file_path} parameter invalid."
        return 2
    fi
}

function check_os_info()
{
    local os_name="$1"
    local os_version_id="$2"
    local cur_os_name
    local cur_os_version_id

    if [ -z "${os_name}" ] || [ -z "${os_version_id}" ];then
        echo "error:inputs null"
        return 1
    fi
    cur_os_name=$(< /etc/os-release grep -E "^NAME" | awk -F "=" '{print $2}' | tr -d '"')
    cur_os_version_id=$(< /etc/os-release grep -E "^VERSION_ID" | awk -F "=" '{print $2}' | tr -d '"')

    # 判断文件内的字段OS_NAME和OS_VERSION_ID是否与当前系统文件/etc/os-release中的NAME和VERSION_ID对应
    if [ "${cur_os_name}" == "${os_name}" ] && [ "${cur_os_version_id}" == "${os_version_id}" ];then
        return 0
    fi

    echo "warn:${cur_os_name} != ${os_name} or ${cur_os_version_id} != ${os_version_id}"
    return 1
}

function is_web_access_enable()
{
    local access_ini_path="/home/data/ies/access_control.ini"
    local section="access_control"
    local key="web_access"
    source /home/data/config/os_cmd.conf
    local web_access_value
    web_access_value=$("${OS_CMD_AWK}" -F '=' '/\['"${section}"'\]/{a=1}a==1&&$1~/'"${key}"'/{print $2;exit}' "${access_ini_path}")
    if [[ "${web_access_value}" == "true" || "${web_access_value}" == " True" ]]; then
      return 0
    fi
    return 1
}