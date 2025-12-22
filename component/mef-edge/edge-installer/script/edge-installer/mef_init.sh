#!/bin/bash
# Copyright (c)  2023. Huawei Technologies Co., Ltd.  All rights reserved.

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
