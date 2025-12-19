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
set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
OPENSOURCE_DIR="${TOP_DIR}/opensource"

function build_opensource() {
    rm -rf "$OPENSOURCE_DIR"
    mkdir -p "$OPENSOURCE_DIR"

    download_opensource "https://github.com/nginx/nginx" "https://gitcode.com/GitHub_Trending/ng/nginx" "release-1.27.4"
    download_opensource "https://github.com/openssl/openssl" "https://gitcode.com/GitHub_Trending/ope/openssl" "openssl-3.0.9"
    download_opensource "https://github.com/PCRE2Project/pcre2" "https://gitcode.com/gh_mirrors/pc/pcre2" "pcre2-10.44"
}

function download_opensource() {
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

function main() {
  build_opensource
}

main