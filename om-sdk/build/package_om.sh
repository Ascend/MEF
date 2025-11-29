#!/bin/bash
# Copyright © Huawei Technologies Co., Ltd. 2020-2025. All rights reserved.
CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$CUR_DIR/..
OUTPUT_DIR="${TOP_DIR}"/output
OUTPUT_PACKAGE_DIR="${TOP_DIR}"/output/package

function create_package_dir()
{
    [ ! -d "${OUTPUT_DIR}"/scripts ] && mkdir "${OUTPUT_DIR}"/scripts
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/config ] && mkdir "${OUTPUT_PACKAGE_DIR}"/config
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/software/om_config ] && mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/om_config
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/software/om_bin ] && mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/om_bin
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/software/om_common ] && mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/om_common
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/software/om_lib ] && mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/om_lib
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/software/RedfishServer ] && mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/RedfishServer
    [ ! -d "${OUTPUT_PACKAGE_DIR}"/web_assets/ ] && mkdir -p "${OUTPUT_PACKAGE_DIR}"/web_assets
}

function package_scripts() {
    local omsdk_tmp_dir="${OUTPUT_DIR}"/tmp_omsdk
    mkdir -p "${omsdk_tmp_dir}"
    local omsdk_tag_gz=$(find "${OUTPUT_DIR}/" -maxdepth 1 -type f | grep "tar.gz$")
    tar -zxf "${omsdk_tag_gz}" -C "${omsdk_tmp_dir}"
    cp "${omsdk_tmp_dir}/scripts/reset_om.sh" "${OUTPUT_DIR}"/scripts
    cp "${TOP_DIR}"/src/app/scripts/integrate_om_files.sh "${OUTPUT_DIR}"/scripts
    cp "${TOP_DIR}"/src/app/scripts/upgrade_om.sh "${OUTPUT_DIR}"/scripts
    cp "${TOP_DIR}"/src/app/scripts/install.sh "${OUTPUT_DIR}"/

    rm -r "${omsdk_tmp_dir}"
}

function package_om_config_dir()
{
    cp -rf "${TOP_DIR}"/config/* "${OUTPUT_PACKAGE_DIR}"/config/
}

function package_software_config()
{
    cp -rf "${TOP_DIR}"/src/app/sys_om/config/* "${OUTPUT_PACKAGE_DIR}"/software/om_config/
}

function package_software_bin() {
    cp -rf "${TOP_DIR}"/src/app/sys_om/om_bin/* "${OUTPUT_PACKAGE_DIR}"/software/om_bin
}

function package_software_common()
{
    cp -rf "${TOP_DIR}"/src/app/sys_om/om_common/* "${OUTPUT_PACKAGE_DIR}"/software/om_common
}

function package_software_lib()
{
    cp -rf "${TOP_DIR}"/src/app/sys_om/om_lib/* "${OUTPUT_PACKAGE_DIR}"/software/om_lib
}

function package_redfish_server()
{
    cp -rf "${TOP_DIR}"/src/app/sys_om/RedfishServer/* "${OUTPUT_PACKAGE_DIR}"/software/RedfishServer
}

function package_web()
{
    cp -rf "${CUR_DIR}"/web_assets/* "${OUTPUT_PACKAGE_DIR}"/web_assets/
}

function main()
{
    create_package_dir
    package_scripts
    package_om_config_dir
    package_software_config
    package_software_bin
    package_software_common
    package_software_lib
    package_redfish_server
    package_web
}

main
exit 0
