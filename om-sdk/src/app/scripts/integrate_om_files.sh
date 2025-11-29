#!/bin/bash
# Copyright © Huawei Technologies Co., Ltd. 2020-2025. All rights reserved.
CUR_DIR=$(dirname $(readlink -f "$0"))
TMP_UNPACK_DIR="${CUR_DIR}"/../temp_unpack
TMP_VERSION_DIR="${CUR_DIR}"/../temp_version
OM_UNPACK_PATH=""
OMSDK_PACKAGE_PATH=""
OM_PACKAGE_PATH=""
VERSION_XML_PATH=""

check_soft_link()
{
    if [[ $# -ne 1 ]]; then
        return 1
    fi

    local input_name="$1"
    local format_name
    local real_name

    format_name=$(realpath -ms "${input_name}")
    real_name=$(readlink -m "${input_name}")
    if [[ "${format_name}" != "${real_name}" ]]; then
        return 1
    fi

    return 0
}

function integrate_om_config_dir()
{
    # 拷贝config/default_capability.json到RedfishServer/config和config
    cp -rf "${TMP_UNPACK_DIR}"/config/default_capability.json "${OM_UNPACK_PATH}"/config/default_capability.json
    cp -rf "${TMP_UNPACK_DIR}"/config/default_capability.json "${OM_UNPACK_PATH}"/software/RedfishServer/config
    cp -rf "${TMP_UNPACK_DIR}"/config/default_policy.dat "${OM_UNPACK_PATH}"/config/default_policy.dat
    cp -rf "${TMP_UNPACK_DIR}"/config/default_policy.dat "${OM_UNPACK_PATH}"/software/sec_agent/default_policy.dat
    cp -rf "${TMP_UNPACK_DIR}"/config/policy.dat "${OM_UNPACK_PATH}"/config/policy.dat
    cp -rf "${TMP_UNPACK_DIR}"/config/policy.dat "${OM_UNPACK_PATH}"/software/sec_agent/policy.dat
    cp -rf "${TMP_UNPACK_DIR}"/config/version.xml "${OM_UNPACK_PATH}"/config/version.xml
    cp -rf "${TMP_UNPACK_DIR}"/config/nginx.conf "${OM_UNPACK_PATH}"/config/nginx.conf
}

function integrate_redfish_server()
{
    local target_redfish_server_dir="${OM_UNPACK_PATH}"/software/RedfishServer
    local om_redfish_server_dir="${TMP_UNPACK_DIR}"/software/RedfishServer
    cp -rf "${om_redfish_server_dir}"/om_event_subscription "${target_redfish_server_dir}"/
    cp -rf "${om_redfish_server_dir}"/om_fd_msg_process "${target_redfish_server_dir}"/
    cp -rf "${om_redfish_server_dir}"/om_kmc_update "${target_redfish_server_dir}"/
    cp -rf "${om_redfish_server_dir}"/om_redfish_db "${target_redfish_server_dir}"/
    cp -rf "${om_redfish_server_dir}"/om_redfish_service/ "${target_redfish_server_dir}"/
    # 覆盖om sdk的路由配置文件
    cp -rf "${om_redfish_server_dir}"/om_routes/route_config.py "${target_redfish_server_dir}"/routes/route_config.py
    cp -rf "${om_redfish_server_dir}"/om_system_service "${target_redfish_server_dir}"/

    # 覆盖扩展配置文件
    cp -rf "${om_redfish_server_dir}"/om_register_blueprint.py "${target_redfish_server_dir}"/
    cp -rf "${om_redfish_server_dir}"/extend_interfaces.py "${target_redfish_server_dir}"/extend_interfaces.py
    cp -rf "${om_redfish_server_dir}"/redfish_extend_funcs.py "${target_redfish_server_dir}"/
    cp -rf "${om_redfish_server_dir}"/fd_extend_interfaces.py "${target_redfish_server_dir}"/fd_extend_interfaces.py
    cp -rf "${om_redfish_server_dir}"/restful_extend_interfaces.py "${target_redfish_server_dir}"/restful_extend_interfaces.py
}

function integrate_software_config()
{
    cp -rf "${TMP_UNPACK_DIR}"/software/om_config/OMClassMap.json "${OM_UNPACK_PATH}"/software/ibma/config/
    cp -rf "${TMP_UNPACK_DIR}"/software/om_config/ensd.conf "${OM_UNPACK_PATH}"/software/ens/conf/ensd.conf
}

function integrate_software_bin() {
    cp -rf "${TMP_UNPACK_DIR}"/software/om_bin/* ${OM_UNPACK_PATH}/software/ibma/bin
}

function integrate_software_common()
{
    local common_dir="${OM_UNPACK_PATH}"/software/ibma/common
    local redfish_dir="${OM_UNPACK_PATH}"/software/RedfishServer/common
    local om_common_dir="${TMP_UNPACK_DIR}"/software/om_common

    # 拷贝OM独有的ResourceDefV1
    mkdir -p "${OM_UNPACK_PATH}"/software/ibma/om_common
    cp -rf "${om_common_dir}"/ResourceDefV1/ "${OM_UNPACK_PATH}"/software/ibma/om_common
    mkdir -p "${OM_UNPACK_PATH}"/software/RedfishServer/om_common
    cp -rf "${om_common_dir}"/ResourceDefV1/ "${OM_UNPACK_PATH}"/software/RedfishServer/om_common

    # 覆盖sdk的产品配置
    cp -rf "${om_common_dir}"/om_constants/om_product_constants.py "${common_dir}"/constants/product_constants.py
    cp -rf "${om_common_dir}"/om_constants/om_product_constants.py "${redfish_dir}"/constants/product_constants.py

    # 拷贝OM独有的MockupData
    local redfish_mockup_data_dir="${redfish_dir}"/MockupData/iBMAServerV1/redfish/v1
    local monitor_mockup_data__dir="${common_dir}"/MockupData/iBMAServerV1/redfish/v1
    local om_mockup_data_dir="${om_common_dir}"/MockupData/iBMAServerV1/redfish/v1

    cp -rf "${om_mockup_data_dir}"/schemastore/en/* "${monitor_mockup_data__dir}"/schemastore/en/
    cp -rf "${om_mockup_data_dir}"/schemastore/en/* "${redfish_mockup_data_dir}"/schemastore/en/
    cp -rf "${om_mockup_data_dir}"/Systems/DigitalWarranty "${monitor_mockup_data__dir}"/Systems/
    cp -rf "${om_mockup_data_dir}"/Systems/DigitalWarranty "${redfish_mockup_data_dir}"/Systems
    cp -rf "${om_mockup_data_dir}"/Systems/index.json "${redfish_mockup_data_dir}"/Systems/
    cp -rf "${om_mockup_data_dir}"/Systems/index.json "${monitor_mockup_data__dir}"/Systems/
    cp -rf "${om_mockup_data_dir}"/EventService "${monitor_mockup_data__dir}"/
    cp -rf "${om_mockup_data_dir}"/EventService "${redfish_mockup_data_dir}"/
    cp -rf "${om_mockup_data_dir}"/index.json "${monitor_mockup_data__dir}"/
    cp -rf "${om_mockup_data_dir}"/index.json "${redfish_mockup_data_dir}"/
    cp -rf "${om_mockup_data_dir}"/om_odata/index.json "${redfish_mockup_data_dir}"/odata/index.json
    cp -rf "${om_mockup_data_dir}"/om_odata/index.json "${monitor_mockup_data__dir}"/odata/index.json
}

function integrate_software_lib()
{
    local om_lib_linux_dir="${TMP_UNPACK_DIR}"/software/om_lib/Linux
    local out_lib_dir="${OM_UNPACK_PATH}"/software/ibma/om_lib/Linux
    mkdir -p "${out_lib_dir}"
    cp -rf "${om_lib_linux_dir}"/om_systems "${out_lib_dir}"/
    cp -rf "${om_lib_linux_dir}/../__init__.py" "${out_lib_dir}"/
}

function integrate_version_xml() {
    cp "${TMP_VERSION_DIR}"/version.xml "${OM_UNPACK_PATH}"
}

function integrate_web() {
    cp -rf "${TMP_UNPACK_DIR}"/web_assets/WhiteboxConfig/* "${OM_UNPACK_PATH}"/software/nginx/html/manager/WhiteboxConfig
    cp -rf "${TMP_UNPACK_DIR}"/web_assets/onlineHelp/* "${OM_UNPACK_PATH}"/software/nginx/html/manager/onlineHelp
}

function remove_not_support_components()
{
    # 删除modules相关的文件
    rm "${OM_UNPACK_PATH}"/software/RedfishServer/system_service/module_views.py
    rm "${OM_UNPACK_PATH}"/software/RedfishServer/routes/systems/modules_route.py
}

function integrate_files() {
    integrate_om_config_dir
    integrate_redfish_server
    integrate_software_config
    integrate_software_bin
    integrate_software_common
    integrate_software_lib
    integrate_version_xml
    remove_not_support_components
    integrate_web
}

function main()
{
    if [[ $# -ne 4 ]]; then
        echo "parameter error."
        return 1
    fi

    # 检查入参路径是否为软连接, 路径不存在时，创建路径
    for dir_path in "$@"; do
        if ! check_soft_link "${dir_path}"; then
            return 1
        fi
    done

    # 代码整合后的存放路径
    OM_UNPACK_PATH="$1"
    # OM SDK tar包路径
    OMSDK_PACKAGE_PATH="$2"
    # A500-A2-om tar包路径
    OM_PACKAGE_PATH="$3"
    VERSION_XML_PATH="$4"

    if [[ ! -d "${OM_UNPACK_PATH}" ]]; then
        mkdir -p "${OM_UNPACK_PATH}"
    fi

    # 创建om包临时解压目录，omsdk包可直接解压到OM_UNPACK_PATH
    [ ! -d "${TMP_UNPACK_DIR}" ] && mkdir -p "${TMP_UNPACK_DIR}"
    # 将version.xml放在一个临时路径里
    [ ! -d "${TMP_VERSION_DIR}" ] && mkdir -p "${TMP_VERSION_DIR}"
    mv "${VERSION_XML_PATH}" "${TMP_VERSION_DIR}"
    # 解压压缩包
    tar --no-same-owner -zxf "${OMSDK_PACKAGE_PATH}" -C "${OM_UNPACK_PATH}" &> /dev/null
    tar --no-same-owner -zxf "${OM_PACKAGE_PATH}" -C "${TMP_UNPACK_DIR}" &> /dev/null
    # 将om解压后的代码移动到OM_UNPACK_PATH对应位置
    integrate_files

    rm -r "${TMP_VERSION_DIR}"
    rm -r "${TMP_UNPACK_DIR}"

}

main "$@"
ret=$?
exit "${ret}"
