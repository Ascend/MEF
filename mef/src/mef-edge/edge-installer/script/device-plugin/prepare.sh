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
CUR_DIR=$(dirname "$(readlink -f "$0")")
source "${CUR_DIR}"/../../edge_installer/script/safe_common.sh

function main()
{
  log_dir=$1
  log_backup_dir=$2
  device_plugin_log_dir="${log_dir}/device_plugin"
  device_plugin_log_backup_dir="${log_backup_dir}/device_plugin"
  device_plugin_log="${log_dir}/device_plugin/device_plugin_run.log"
  create_log_dir "$log_dir" 755
  create_log_dir "$log_backup_dir" 755
  create_log_dir "$device_plugin_log_dir" 750
  create_log_dir "$device_plugin_log_backup_dir" 750
  create_log "$device_plugin_log" 640
}

main "$@"
exit 0