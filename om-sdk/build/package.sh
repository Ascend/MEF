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

CURDIR=$(dirname $(readlink -f "$0"))
SCRIPT_NAME=$(basename $0)
ROOT_PATH=$(readlink -f $CURDIR/../)
OUTPUT_PATH="$ROOT_PATH/output"

function make_zip_package()
{
    #directory need to create
    local atlasedge=$(ls $OUTPUT_PATH/*-om*.tar.gz)
    local atlasedge1=${atlasedge##*/}
    local atlasedge_release=${atlasedge1%.*.*}
    local chmod_prepare_script="${CURDIR}/chmod_prepare.sh"
    echo $atlasedge_release


    cd ${OUTPUT_PATH}

    if [ -z "${atlasedge_release}" ]; then
        echo "get package name failed"
        return -1
    fi

    local package_file=${OUTPUT_PATH}/package
    [ ! -d "${package_file}" ] && mkdir ${package_file}
    cp -rfd ${OUTPUT_PATH}/crldata.crl $OUTPUT_PATH/${atlasedge_release}.tar.gz.crl
    mv $atlasedge_release.* $package_file
    cp -rfd ${OUTPUT_PATH}/crldata.crl ${OUTPUT_PATH}/version.xml.crl
    cp -rfd ${OUTPUT_PATH}/version.xml.crl ${package_file}/
    cp -rfd ${OUTPUT_PATH}/version.xml.cms ${package_file}/
    cp -rfd ${OUTPUT_PATH}/version.xml ${package_file}/

    sh "${chmod_prepare_script}" "${package_file}"

    cd $package_file
    zip $atlasedge_release.zip \
        $atlasedge_release.tar.gz \
        $atlasedge_release.tar.gz.cms \
        $atlasedge_release.tar.gz.crl \
        version.xml \
        version.xml.cms \
        version.xml.crl
    mv $package_file/$atlasedge_release.zip ${OUTPUT_PATH}/$atlasedge_release.zip
    echo "zip $atlasedge_release success!"
    [[ -d "${package_file}" ]] && sudo rm -rf $package_file

    return 0
}

function main()
{
    make_zip_package
    return 0
}

echo "begin to execute $SCRIPT_NAME"
main $*;ret=$?
echo "finish execute $SCRIPT_NAME, result is ${ret}!"
exit $ret
