#!/bin/bash
# Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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