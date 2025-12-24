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

umask 0027

function recover_log() {
    local dir_path=$1

    if [[ -d "${dir_path}" ]]; then
      echo "log_dir exists, no need recover"
      return
    fi

    echo "log_dir doesn't exist, start to recover log"
    local current_dir=$(dirname "$(readlink -f "$0")")
    local edge_installer_dir=$(readlink -f "${current_dir}/../bin")
    "${edge_installer_dir}/innerctl" recover_log
    echo "recover log success"
}

function main() {
    local dir_path=$1

    recover_log "$dir_path"
}

main "$@"
exit 0
