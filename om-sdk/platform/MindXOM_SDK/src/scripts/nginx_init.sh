#!/bin/bash
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

function check_nginx_is_up()
{
    # 等待端口拉起次数，端口拉起表示nginx启动成功
    local wait_times=0
    while [[ $(netstat -atnp | grep nginx | awk -F " " '{print $4}' | awk -F ":" '{print $2}' | head -1) -eq "" ]]
    do
        if [[ "${wait_times}" -ge 8 ]]
        then
            return 1
        fi
        wait_times=$((wait_times + 1))
        sleep 0.5
    done
    return 0
}


function start_nginx()
{
    if ! "${OM_WORK_DIR}"/scripts/stop_nginx.sh; then
        # 启动之前先停止，停止失败记录日志，不影响后续启动流程
        logger_warn "stop nginx failed"
    fi
    local wait_times=0
    "${OM_WORK_DIR}"/scripts/start_nginx.sh &
    while ! check_nginx_is_up
    do
        if [[ "${wait_times}" -ge 5 ]]
        then
            break
        fi
        wait_times=$((wait_times + 1))
        logger_info "nginx start failed, will retry, times: ${wait_times}."

        if ! pgrep -f "${OM_WORK_DIR}"/scripts/start_nginx.sh > /dev/null 2>&1
        then
            "${OM_WORK_DIR}"/scripts/start_nginx.sh &
        fi
    done
    logger_info "start_nginx func done"
}

function main() {
  CURRENT_PATH=$(cd "$(dirname "$0")" || exit 3; pwd)
  source "${CURRENT_PATH}"/log_print.sh
  init_log_path "/var/plog/ibma_edge" "om_scripts_run.log"
  logger_info "start the nginx init!"
  if [[ ! -d "/home/log/plog/web_edge" ]]; then
      mkdir -p "/home/log/plog/web_edge"
      chmod 750 "/home/log/plog/web_edge"
  fi

  if [ ! -f /home/data/ies/access_control.ini ];then
    if [ -L /home/data/ies/access_control.ini ]; then
      logger_warn "/home/data/ies/access_control.ini is soft link, unlink it"
      unlink /home/data/ies/access_control.ini
    fi
    logger_info "default access_control does not exit, start copy."
    cp -rpf "${OM_WORK_DIR}"/software/ibma/lib/Linux/config/access_control.ini /home/data/ies/access_control.ini
  fi

  source /usr/local/mindx/MindXOM/scripts/safe_common.sh
  if ! is_web_access_enable; then
    return 0
  fi

  # Check and refresh nginx config
  export PYTHONPATH="${OM_WORK_DIR}"/software/ibma:"${OM_WORK_DIR}"/software/ibma/opensource/python
  if ! python3 "${OM_WORK_DIR}"/software/ibma/lib/Linux/systems/nic/config_web_ip.py "${OM_WORK_DIR}"/config/nginx.conf "Start"; then
      logger_error "Check and refresh nginx config failed. please check /var/plog/manager/manager_run.log."
      return 1
  fi

  start_nginx

  logger_info "finish the nginx init!"
  return 0
}

main