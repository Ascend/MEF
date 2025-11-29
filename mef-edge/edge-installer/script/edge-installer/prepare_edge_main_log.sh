#!/bin/bash
# Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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