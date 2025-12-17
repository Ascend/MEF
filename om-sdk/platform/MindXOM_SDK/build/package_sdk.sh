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

CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR="${CUR_DIR}"/..
OUTPUT_PACKAGE_DIR="${TOP_DIR}"/output/package
OM_WORK_DIR=/usr/local/mindx/MindXOM

function create_package_dir()
{
    mkdir -p "${OUTPUT_PACKAGE_DIR}"
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/bin
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/config
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/lib
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software
}

function package_script_to_tools_dir()
{
    cp -rf "${TOP_DIR}"/tools "${OUTPUT_PACKAGE_DIR}"/
}

function package_config_dir()
{
    cp -rf "${TOP_DIR}"/config/* "${OUTPUT_PACKAGE_DIR}"/config
}

function package_lib_dir()
{
    cp -f "${TOP_DIR}"/output/lib/libcertmanage.so "${OUTPUT_PACKAGE_DIR}"/lib
    cp -f "${TOP_DIR}"/output/lib/libssl.so* "${OUTPUT_PACKAGE_DIR}"/lib
    cp -f "${TOP_DIR}"/output/lib/libcrypto.so* "${OUTPUT_PACKAGE_DIR}"/lib
    cp -f "${TOP_DIR}"/output/lib/libsecurec.so "${OUTPUT_PACKAGE_DIR}"/lib
    cp -f "${TOP_DIR}"/output/lib/libcommon.so "${OUTPUT_PACKAGE_DIR}"/lib
    cp -f "${TOP_DIR}"/output/lib/liblpeblock.so "${OUTPUT_PACKAGE_DIR}"/lib
}

function package_scripts_dir()
{
    cp -rf "${TOP_DIR}"/src/scripts "${OUTPUT_PACKAGE_DIR}"/
    mv "${OUTPUT_PACKAGE_DIR}"/scripts/install.sh "${OUTPUT_PACKAGE_DIR}"/
    mv "${OUTPUT_PACKAGE_DIR}"/scripts/uninstall.sh "${OUTPUT_PACKAGE_DIR}"/
    mv "${OUTPUT_PACKAGE_DIR}"/scripts/upgrade.sh "${OUTPUT_PACKAGE_DIR}"/
}

function package_redfishserver_dir()
{
    # pakeage RedfishServer和common到software/RedfishServe目录下
    local redfish_server_dir="${OUTPUT_PACKAGE_DIR}"/software/RedfishServer
    mkdir -p "${redfish_server_dir}"
    cp -rf "${TOP_DIR}"/src/app/sys/RedfishServer/* "${redfish_server_dir}"/
    cp -rf "${TOP_DIR}"/src/app/sys/common "${redfish_server_dir}"/

    # 拷贝python到lib/python
    local redfish_server_lib_dir="${redfish_server_dir}"/lib
    mkdir -p "${redfish_server_lib_dir}"/python
    cp -rf "${TOP_DIR}"/output/opensource/python/* "${redfish_server_lib_dir}"/python/

    # 拷贝log_print.sh等执行脚本到RedfishServer目录
    cp -f "${TOP_DIR}"/src/scripts/log_print.sh "${redfish_server_dir}"

    # 拷贝config/iBMA.ini到RedfishServer
    local redfish_config_dir=$redfish_server_dir/config
    mkdir -p "${redfish_config_dir}"
    cp -f "${TOP_DIR}"/src/app/sys/config/iBMA.ini "${redfish_config_dir}"
    # 拷贝config/default_capability.json到RedfishServer/config
    cp -f "${TOP_DIR}"/config/default_capability.json "${redfish_config_dir}"
    # 拷贝config/not_support_component_config.json到RedfishServer/config
    cp -f "${TOP_DIR}"/config/not_support_component_config.json "${redfish_config_dir}"
    # 拷贝ibma_edge_service.ini到RedfishServer/config，方便Redfish读取FD的域名对应IP信息
    cp -f "${TOP_DIR}"/src/app/sys/lib/Linux/config/ibma_edge_service.ini "${redfish_config_dir}"
    # 拷贝alarm_info_en.json到RedfishServer/config
    cp -f "${TOP_DIR}"/config/alarm_info_en.json "${redfish_config_dir}"

    # 拷贝c库到lib/c
    local lib_c_dir="${redfish_server_lib_dir}"/c
    mkdir -p "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libcertmanage.so "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libsecurec.so "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libcommon.so "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libcrypto.so* "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libssl.so* "${lib_c_dir}"
}

function package_software_dir()
{
    cp -f "${TOP_DIR}"/config/version.xml "${OUTPUT_PACKAGE_DIR}"
    cp -f "${TOP_DIR}"/config/version.xml "${TOP_DIR}"/output/version.xml

    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/ibma
    cp -rf "${TOP_DIR}"/src/service_main "${OUTPUT_PACKAGE_DIR}"/software/
    cp -rf "${TOP_DIR}"/src/app/sys/bin "${OUTPUT_PACKAGE_DIR}"/software/ibma
    cp -rf "${TOP_DIR}"/src/app/sys/common "${OUTPUT_PACKAGE_DIR}"/software/ibma
    cp -rf "${TOP_DIR}"/src/app/sys/devm "${OUTPUT_PACKAGE_DIR}"/software/ibma
    cp -rf "${TOP_DIR}"/src/app/sys/monitor_db "${OUTPUT_PACKAGE_DIR}"/software/ibma
    cp -rf "${TOP_DIR}"/src/app/sys/lib "${OUTPUT_PACKAGE_DIR}"/software/ibma
    cp -rf "${TOP_DIR}"/src/app/sys/monitor_kmc "${OUTPUT_PACKAGE_DIR}"/software/ibma

    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/ibma/opensource/python
    cp -rf "${TOP_DIR}"/output/opensource/python/* "${OUTPUT_PACKAGE_DIR}"/software/ibma/opensource/python

    cp -rf "${TOP_DIR}"/src/app/sys/config "${OUTPUT_PACKAGE_DIR}"/software/ibma

    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/
    # package ens
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/ens/bin
    cp -f "${TOP_DIR}"/src/cpp/ens/start_ensd.sh "${OUTPUT_PACKAGE_DIR}"/software/ens/bin
    cp -f "${TOP_DIR}"/src/build/cpp/ens/ensd "${OUTPUT_PACKAGE_DIR}"/software/ens/bin
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/ens/conf
    cp -f "${TOP_DIR}"/src/cpp/ens/ensd.conf "${OUTPUT_PACKAGE_DIR}"/software/ens/conf
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/ens/lib
    cp -f "${TOP_DIR}"/output/lib/libbase.so "${OUTPUT_PACKAGE_DIR}"/software/ens/lib
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/ens/modules
    cp -f "${TOP_DIR}"/output/lib/libalarm_process.so "${OUTPUT_PACKAGE_DIR}"/software/ens/modules
    cp -f "${TOP_DIR}"/output/lib/libdevm.so "${OUTPUT_PACKAGE_DIR}"/software/ens/modules
    cp -f "${TOP_DIR}"/output/lib/libfault_check.so "${OUTPUT_PACKAGE_DIR}"/software/ens/modules
    cp -f "${TOP_DIR}"/output/lib/libextend_alarm.so "${OUTPUT_PACKAGE_DIR}"/software/ens/modules
}

function package_nginx_dir() {
    # package nginx
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/
    cp -rf "${OM_WORK_DIR}"/software/nginx "${OUTPUT_PACKAGE_DIR}"/software/
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/nginx/html/manager
    cp -rf "${TOP_DIR}"/src/app/web/* "${OUTPUT_PACKAGE_DIR}"/software/nginx/html/manager
    rm -rf "${OUTPUT_PACKAGE_DIR}"/software/nginx/html/manager/onlinehelp
    # 拷贝python公共代码
    mkdir -p "${OUTPUT_PACKAGE_DIR}"/software/nginx/python
    cp -rf "${TOP_DIR}"/src/app/sys/common "${OUTPUT_PACKAGE_DIR}"/software/nginx/python
    rm -rf "${OUTPUT_PACKAGE_DIR}"/software/nginx/python/MockupData
    rm -rf "${OUTPUT_PACKAGE_DIR}"/software/nginx/python/ResourceDefV1
    mv "${OUTPUT_PACKAGE_DIR}"/scripts/python/start_nginx.py "${OUTPUT_PACKAGE_DIR}"/software/nginx/python
    # 拷贝c库
    local lib_c_dir="${OUTPUT_PACKAGE_DIR}"/software/nginx/lib
    mkdir -p "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libsecurec.so "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libssl.so* "${lib_c_dir}"
    cp -f "${TOP_DIR}"/output/lib/libcrypto.so* "${lib_c_dir}"
}

function main()
{
    create_package_dir
    package_script_to_tools_dir
    package_config_dir
    package_lib_dir
    package_scripts_dir
    package_software_dir
    package_nginx_dir
    package_redfishserver_dir
}

main
exit 0
