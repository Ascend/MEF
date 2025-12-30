#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MEF is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

#***********************************************
#  Description: 校验chmod中path是否是软连接
#    Parameter: 1、mode
#               2、path
#               3、-R
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
function safe_chmod() {
    if [ $# -lt 2 ] || [ $# -gt 3 ] || [[ $# -gt 2  && "$3" != "-R" ]]; then
       echo "safe chmod parameter error"
       return 1
    fi

    local mode="$1"

    local path
    path=$(realpath -s "$2")

    # path is link
    if [ -L "${path}" ]; then
        echo "${path} does not comply with security rules." && return 1
    fi

    shift 2

    chmod "${mode}" "${path}" "$@"
    return 0
}

#***********************************************
#  Description: 校验拷贝文件的src,dst是否是软连接
#    Parameter: 1、srcpath
#               2、dstpath
#        Input: NA
#       Output: NA
# Return Value: 0 -- 参数校验通过
#               1 -- 参数校验不通过
#      Cattion: NA
#***********************************************
function safe_cp() {
    if [ $# -lt 2 ]; then
        echo "safe cp parameter error"
        return 1
    fi
    local srcpath
    local dstpath

    srcpath=$(realpath -s "$1")
    dstpath=$(realpath -s "$2")

    # srcpath is link
    if [ -L "${srcpath}" ]; then
        echo "${srcpath} does not comply with security rules." && return 1
    fi

    # dstpath is link
    if [ -L "${dstpath}" ]; then
        echo "${dstpath} does not comply with security rules." && return 1
    fi

    cp "$@"
    return 0
}

#***********************************************
#  Description: 创建日志文件
#    Parameter: 1、log_path
#               2、log_mode
#        Input: NA
#       Output: NA
# Return Value: NA
#      Cattion: NA
#***********************************************
function create_log()
{
  local log_path=$1
  local log_mode=640
  if [ $# -gt 2 ]; then
     log_mode=$2
  fi

  if [[ -L ${log_path} ]]; then
      unlink "${log_path}"
  elif [[ -d ${log_path} ]]; then
      rm -r "${log_path}"
  fi

  if [[ ! -f ${log_path} ]]; then
      touch "${log_path}"
  fi

  safe_chmod "${log_mode}" "${log_path}"
}

#***********************************************
#  Description: 创建日志文件目录
#    Parameter: 1、dir_path
#               2、dir_mode
#               3、dir_owner
#        Input: NA
#       Output: NA
# Return Value: NA
#      Cattion: NA
#***********************************************
function create_log_dir()
{
    local dir_path=$1
    local dir_mode=$2
    local dir_owner="root"
    if [ $# -gt 2 ]; then
        dir_owner=$3
    fi

    if [[ -L "${dir_path}" ]]; then
        unlink "${dir_path}"
    elif [[ -f "${dir_path}" ]]; then
        unlink "${dir_path}"
    fi

    if [[ ! -d "${dir_path}" ]]; then
      mkdir -p "${dir_path}"
    fi

    safe_chmod "${dir_mode}" "${dir_path}"
    chown "${dir_owner}":"${dir_owner}" "${dir_path}"
}