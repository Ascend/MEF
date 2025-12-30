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

umask 0022
# to protect mkdir cmd, if a file is not softlink or normal file, when makedir failed, mod and owner will be changed
# by root
set -e

CUR_DIR=$(dirname "$(readlink -f "$0")")
source "${CUR_DIR}"/safe_common.sh

function main()
{
  log_dir=$1
  log_backup_dir=$2
  log_root_dir="$(dirname "${log_dir}")"
  log_backup_root_dir="$(dirname "${log_backup_dir}")"
  edge_main_log_dir="${log_dir}/edge_main"
  edge_main_log_backup_dir="${log_backup_dir}/edge_main"
  create_log_dir "${log_root_dir}" 755 "root"
  create_log_dir "${log_backup_root_dir}" 755 "root"
  create_log_dir "${log_dir}" 755 "root"
  create_log_dir "${log_backup_dir}" 755 "root"
  create_log_dir "${edge_main_log_dir}" 750 "MEFEdge"
  create_log_dir "${edge_main_log_backup_dir}" 750 "MEFEdge"
}

main "$@"
exit 0