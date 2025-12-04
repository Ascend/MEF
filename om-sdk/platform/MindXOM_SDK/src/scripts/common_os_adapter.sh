#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
# Description: 通用OS适配盘符处理脚本
#***********************************************
# Description: 获取系统盘的主分区的盘符路径
# Parameter:
#   input:  NA
#   output: 系统盘的主分区的盘符路径
# Return Value : NA
#***********************************************
function get_sys_main_dev_path()
{
    local sys_main_dev=""
    sys_main_dev=$(grep "root=" /proc/cmdline | sed 's/.*\sroot=//g' | awk -F" " '{print $1}')
    if [[ "${sys_main_dev}" =~ ^PARTUUID=[0-9a-f\-]{36,36}$ ]]; then
        sys_main_dev=$(blkid -t "${sys_main_dev}" | awk -F: '{print $1}')
    fi
    echo "${sys_main_dev}"
}

#***********************************************
# Description: 获取系统盘的主分区的分区标签
# Parameter:
#   input:  NA
#   output: 当前系统盘的主分区的分区标签
# Return Value : NA
#***********************************************
function get_sys_main_dev_label()
{
    local sys_main_dev=""
    sys_main_dev=$(get_sys_main_dev_path)
    local part_label=""
    part_label=$(blkid "${sys_main_dev}" | sed 's/.*\sPARTLABEL=//g' | awk -F" " '{print $1}' | tr -d '"')
    echo "${part_label}"
}

#***********************************************
# Description: 判断当前系统是否从emmc启动
# Parameter:
#   input:  NA
#   output: NA
# Return Value: 0 -- 是
#               1 -- m.2
#***********************************************
function check_boot_from_emmc()
{
    local sys_main_dev=""
    sys_main_dev=$(get_sys_main_dev_path)
    if [[ "${sys_main_dev}" =~ ^/dev/mmcblk0p[0-9]+$ ]]; then
        return 0
    fi
    return 1
}

#***********************************************
# Description: 获取系统主分区的分区号，只可能是2或者3
# Parameter:
#   input:  NA
#   output: 当前系统盘的分区号
# Return Value : NA
#***********************************************
function get_sys_main_dev_partition_num()
{
    local sys_main_dev_label=""
    sys_main_dev_label=$(get_sys_main_dev_label)
    part_num=${sys_main_dev_label: 1}
    echo "${part_num}"
}

#***********************************************
# Description: 获取系统备区的分区号，只可能是2或者3
# Parameter:
#   input:  NA
#   output: 当前系统盘的分区号
# Return Value : NA
#***********************************************
function get_sys_back_dev_partition_num()
{
    local sys_main_dev_label=""
    sys_main_dev_label=$(get_sys_main_dev_label)
    part_num=${sys_main_dev_label: 1}
    if [[ "${part_num}" == "2" ]]; then
        echo "3"
    else
        echo "2"
    fi
}

#***********************************************
# Description: 获取系统盘符路径名
# Parameter:
#   input:  NA
#   output: 当前系统盘符路径名
# Return Value : NA
#***********************************************
function get_sys_dev_path()
{
    local main_dev=""
    main_dev=$(get_sys_main_dev_path)
    if check_boot_from_emmc; then
        echo "${main_dev: 0: -2}"
    else
        echo "${main_dev: 0: -1}"
    fi
}

#***********************************************
# Description: 获取系统备区盘符路径
# Parameter:
#   input:  NA
#   output: 当前系统备区盘符路径
# Return Value : NA
#***********************************************
function get_sys_back_dev()
{
    local sys_dev_path=""
    sys_dev_path=$(get_sys_dev_path)
    local sys_back_part_num=""
    sys_back_part_num=$(get_sys_back_dev_partition_num)
    if check_boot_from_emmc; then
        echo "${sys_dev_path}p${sys_back_part_num}"
    else
        echo "${sys_dev_path}${sys_back_part_num}"
    fi
}

#***********************************************
# Description: 获取系统分区盘符路径带后缀p字符
# Parameter:
#   input:  NA
#   output: 当前系统分区盘符路径
# Return Value : NA
#***********************************************
function get_sys_dev_path_suffix()
{
    local main_dev=""
    main_dev=$(get_sys_main_dev_path)
    echo "${main_dev: 0: -1}"
}