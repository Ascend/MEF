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

set -e

CURDIR=$(dirname $(readlink -f "$0"))
SCRIPT_NAME=$(basename $0)
ROOT_PATH=$(readlink -f $CURDIR/../)
OUTPUT_PATH="$ROOT_PATH/output"


function generate_2_vercfg_xml() {
    local vercfg_xml_path="$1"
    local package_dir="${OUTPUT_PATH}/package"

    # 生成vercfg.xml
    touch "${vercfg_xml_path}"
    chmod 600 "${vercfg_xml_path}"
    echo -e '<?xml version="1.0" encoding="utf-8"?>\n<Package>\n</Package>' > "${vercfg_xml_path}"

    # 计算OM驱动包sha256值，写入vercfg.xml/
    local om_tar_gz="$(find "${package_dir}" -maxdepth 1 -type f | grep "Ascend-mindxedge-om.\{1,\}.tar.gz$")"
    sed -i "3 i \ \t<File>\n\t\t<FilePath>$(basename "${om_tar_gz}")</FilePath>\n\t\t<SHAValue>$(sha256sum "${om_tar_gz}" | awk '{print $1}')</SHAValue>\n\t</File>" "${vercfg_xml_path}"

    # 计算version.xml的sha256值，写入vercfg.xml
    local version_xml_path="${package_dir}/version.xml"
    sed -i "3 i \ \t<File>\n\t\t<FilePath>$(basename "${version_xml_path}")</FilePath>\n\t\t<SHAValue>$(sha256sum "${version_xml_path}" | awk '{print $1}')</SHAValue>\n\t</File>" "${vercfg_xml_path}"

    cat "${vercfg_xml_path}"

}


function make_zip_package()
{
    #directory need to create
    local om_tar_gz="$(find "${OUTPUT_PATH}" -maxdepth 1 -type f | grep "Ascend-mindxedge-om.\{1,\}.tar.gz$")"
    local om_basename="$(basename "${om_tar_gz}" ".tar.gz")"
    local package_dir="${OUTPUT_PATH}/package"
    rm -rf "${package_dir}"
    mkdir -p "${package_dir}"

    for file in $(find "${OUTPUT_PATH}" -maxdepth 1 -type f | grep "Ascend-mindxedge-om.\{1,\}.tar.gz"); do
        mv "${file}" "${package_dir}"
    done
    mv "${OUTPUT_PATH}/version.xml" "${package_dir}"

    # 生成并签名vercfg.xml
    local vercfg_xml_path="${OUTPUT_PATH}/vercfg.xml"
    generate_2_vercfg_xml "${vercfg_xml_path}"

    # 打包package目录下的文件
    for file in $(find "${OUTPUT_PATH}" -maxdepth 1 -type f | grep "vercfg.xml"); do
        mv "${file}" "${package_dir}"
    done
    cd "${package_dir}"
    zip "${om_basename}.zip" *
    mv "${package_dir}/${om_basename}.zip" "${OUTPUT_PATH}/${om_basename}.zip"
    [[ -d "${package_dir}" ]] && rm -rf "${package_dir}"

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
