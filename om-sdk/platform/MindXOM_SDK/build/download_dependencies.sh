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

set -e

TOP_DIR="$(dirname "$(dirname "$(readlink -f $0)")")"
OPENSOURCE_DIR="${TOP_DIR}/opensource"

function download_each() {
    local repo_url="$1"
    local repo_mirror_url="$2"
    local tag="$3"

    local repo_dir=$(echo "$repo_url" | awk -F'/' '{print $5}')
    if [[ "$repo_mirror_url" != "" ]]; then
      repo_url="$repo_mirror_url"
    fi

    pushd "$OPENSOURCE_DIR"

    rm -rf "$repo_dir"
    git clone  --depth 1 --branch "$tag" "$repo_url"
    if [ $# -eq 4 ] && [ "$repo_dir" != "$target_dir" ] ; then
        local target_dir="$4"
        rm -rf "$target_dir"
        mv "$repo_dir" "$target_dir"
    fi

    popd
}

function download_all() {
    rm -rf "$OPENSOURCE_DIR"
    mkdir -p "$OPENSOURCE_DIR"
    
    # python dependencies
    download_each "https://github.com/urllib3/urllib3" "https://gitcode.com/gh_mirrors/ur/urllib3" "1.26.19"
    download_each "https://github.com/pallets/click" "https://gitcode.com/gh_mirrors/cl/click" "8.1.3"
    download_each "https://github.com/pallets/jinja" "https://gitcode.com/gh_mirrors/ji/jinja" "3.1.6"
    download_each "https://github.com/pallets/werkzeug" "https://gitcode.com/gh_mirrors/we/werkzeug" "2.2.3"
    download_each "https://github.com/pallets/flask" "https://gitcode.com/gh_mirrors/fl/flask" "2.2.5"
    download_each "https://github.com/sqlalchemy/sqlalchemy" "https://gitcode.com/gh_mirrors/sq/sqlalchemy" "rel_1_4_22"
    download_each "https://github.com/python-websockets/websockets" "https://gitcode.com/gh_mirrors/we/websockets" "10.3"
    # C/C++ dependencies
    download_each "https://github.com/PCRE2Project/pcre2" "https://gitcode.com/gh_mirrors/pc/pcre2" "pcre2-10.44"
    download_each "https://github.com/openssl/openssl" "https://gitcode.com/gh_mirrors/ope/openssl" "openssl-3.0.7"
    download_each "https://github.com/nginx/nginx" "https://gitcode.com/gh_mirrors/ng/nginx" "release-1.27.1"
    download_each "https://gitee.com/openeuler/libboundscheck" "" "v1.1.16"
    download_each "https://github.com/google/googletest" "https://gitcode.com/gh_mirrors/go/googletest" "v1.17.0"
}

download_all
