#!/bin/bash
# Copyright (c)  2022. Huawei Technologies Co., Ltd.  All rights reserved.

umask 0022
CUR_DIR=$(dirname "$(readlink -f "$0")")
CORE_PATH=$(readlink -f "${CUR_DIR}"/../)
PAUSE_TAR_PATH="${CORE_PATH}/bin/pause.tar.gz"
source "${CUR_DIR}"/../../edge_installer/script/safe_common.sh

function ensure_pause_image()
{
  if ! docker load < "${PAUSE_TAR_PATH}";then
      echo "warning:load pause.tar.gz failed"
  else
      echo "load pause.tar.gz success"
  fi
  docker image prune -f
}

function ensure_resolv()
{
    local resolv_conf="/etc/resolv.conf"
    local resolv_mode="644"

     if [[ ! -e "$resolv_conf" ]]; then
        touch "$resolv_conf"
        if ! safe_chmod "$resolv_mode" "$resolv_conf"; then
            echo "chmod $resolv_mode failed"
            return 1
        fi
        echo "Created an empty $resolv_conf"
    fi

}

function main()
{
  log_dir=$1
  log_backup_dir=$2
  edge_core_log_dir="${log_dir}/edge_core"
  edge_core_log_backup_dir="${log_backup_dir}/edge_core"
  edge_core_log="${log_dir}/edge_core/edge_core_run.log"
  create_log_dir "$log_dir" 755
  create_log_dir "$log_backup_dir" 755
  create_log_dir "$edge_core_log_dir" 750
  create_log_dir "$edge_core_log_backup_dir" 750
  create_log "$edge_core_log" 640
  ensure_resolv
  ensure_pause_image

  iptables -D INPUT -p tcp -j PORT-LIMIT-RULE || true
  iptables -F PORT-LIMIT-RULE || true
  iptables -X PORT-LIMIT-RULE || true
  iptables -t filter -N PORT-LIMIT-RULE
  iptables -I INPUT -p tcp -j PORT-LIMIT-RULE
}

main "$@"
exit 0