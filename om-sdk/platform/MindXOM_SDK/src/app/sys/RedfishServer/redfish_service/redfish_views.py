# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

"""
功    能：Redfish公共固定资源 URL处理模块
修改记录：2022-11-1 创建
"""
import enum
import json
import os

from flask import Blueprint

from common.constants import base_constants
from common.constants.base_constants import CommonConstants
from common.constants.product_constants import SERVICE_ROOT
from common.file_utils import FileCheck
from common.log.logger import run_log
from common.utils.app_common_method import AppCommonMethod
from extend_interfaces import EXTEND_REDFISH_SCHEMA
from ibma_redfish_globals import RedfishGlobals
from token_auth import get_privilege_auth

privilege_auth = get_privilege_auth()

public_bp = Blueprint("public", __name__, url_prefix="/redfish")

PRJ_DIR: str = AppCommonMethod.get_project_absolute_path()
MOCKUP_PATH: str = os.path.join(PRJ_DIR, "common/MockupData/iBMAServerV1")


@public_bp.route("", methods=["GET"])
@privilege_auth.token_required
def rf_version_service():
    """ 功能描述： 查询当前使用的Redfish协议的版本号 """
    input_err_info = "Query redfish version info failed."
    try:
        return SERVICE_ROOT.publicService.get_resource()
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_err_info
        }
        return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)


@public_bp.route("/v1", methods=["GET"])
@privilege_auth.token_required
def rf_service_collection():
    """ 功能描述： 查询服务器当前根服务资源 """
    input_err_info = "Query service collections info failed."
    try:
        return SERVICE_ROOT.publicService.service_resource.get_resource()
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_err_info
        }
        return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)


@public_bp.route("/v1/$metadata", methods=["GET"])
@privilege_auth.token_required
def rf_metadata_service():
    """ 功能描述： 查询Redfish规范里的元数据文档 """
    input_err_info = "Query metadata info failed."
    metadata_path = os.path.join(MOCKUP_PATH, "redfish/v1/metadata/index.xml")
    try:
        res = FileCheck.check_xml_file_valid(metadata_path)
        if not res:
            raise ValueError("%s invalid :%s", metadata_path, res.error)
        with open(metadata_path, encoding="utf8") as res_file:
            response = res_file.read(CommonConstants.OM_READ_FILE_MAX_SIZE_BYTES)
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_err_info
        }
        return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)
    return response


@public_bp.route("/v1/JSONSchemas", methods=["GET"])
@privilege_auth.token_required
def rf_schema_collection():
    """ 功能描述： 查询服务器当前所有资源的Schema文件 """
    input_err_info = "Query schema collections info failed."
    # 替代模板中相应字段，完成处理
    members = []
    for member in SchemaCollction:
        members.append({"@odata.id": f"redfish/v1/JSONSchemas/{member.value}"})
    # 扩展schema文件
    members.extend([{"@odata.id": f"redfish/v1/JSONSchemas/{item}"} for item in EXTEND_REDFISH_SCHEMA])
    ret_dict = {"Members@odata.count": len(members), "Members": members}
    try:
        resp_json = json.loads(SERVICE_ROOT.publicService.schema_collection.get_resource())
        RedfishGlobals.replace_kv_list(resp_json, ret_dict)
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_err_info
        }
        return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)
    return json.dumps(resp_json)


@public_bp.route("/v1/JSONSchemas/<member_id>", methods=["GET"])
@privilege_auth.token_required
def rf_schema_info(member_id):
    """ 功能描述： 查询单个Schema文件归档地址 """
    input_err_info = "Query schema info failed."
    # 检查外部参数
    check_ret = check_external_parameter(member_id)
    if check_ret is not None:
        return check_ret
    try:
        resp_json = json.loads(SERVICE_ROOT.publicService.schema_resource.get_resource())
        schema_replace_odata_id(resp_json, member_id)
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_err_info
        }
        return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)
    return json.dumps(resp_json)


@public_bp.route("/v1/odata", methods=['GET'])
@privilege_auth.token_required
def rf_odata_service():
    """ 功能描述： 查询OData服务文档 """
    input_err_info = "Query oData info failed."
    try:
        return SERVICE_ROOT.publicService.odata_resource.get_resource()
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = {
            "status": base_constants.CommonConstants.ERR_CODE_400,
            "message": input_err_info
        }
        return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)


def check_external_parameter(para_data):
    """ 检查外部参数 """
    input_err_info = "Parameter is invalid."
    schema_collection = [member.value for member in SchemaCollction]
    # 扩展schema文件
    schema_collection.extend(EXTEND_REDFISH_SCHEMA)
    if para_data in schema_collection:
        return None
    run_log.error("%s", input_err_info)
    ret_dict = {
        "status": base_constants.CommonConstants.ERR_CODE_400,
        "message": input_err_info
    }
    return RedfishGlobals.return_error_info_message(ret_dict, base_constants.CommonConstants.ERR_GENERAL_INFO)


def schema_replace_odata_id(resp_json, schema_id):
    """
    功能描述：替换模板中的ID
    参数：resp_json JSON 模板
         schema_id 替换的ID
    异常描述：无
    修改记录：2022-12-16 创建
    """
    schema_id_no_version = schema_id
    if "." in schema_id:
        schema_id_no_version = schema_id.split(".")[0]
    resp_json["Id"] = schema_id
    resp_json["@odata.id"] = resp_json["@odata.id"].replace("oDataID1", schema_id)
    resp_json["Description"] = resp_json["Description"].replace('oDataID2', schema_id_no_version)
    resp_json["Name"] = resp_json["Name"].replace('oDataID2', schema_id_no_version)
    resp_json["Schema"] = resp_json["Schema"].replace('oDataID1', schema_id)
    resp_json["Schema"] = resp_json["Schema"].replace('oDataID2', schema_id_no_version)
    resp_json["Location"][0]["Uri"] = resp_json["Location"][0]["Uri"].replace('oDataID1', schema_id)
    resp_json["Location"][0]["PublicationUri"] = resp_json["Location"][0]["PublicationUri"].replace('oDataID1',
                                                                                                    schema_id)
    if schema_id.startswith("MindXEdge"):
        del resp_json["Location"][0]["PublicationUri"]
    else:
        del resp_json["Location"][0]["Uri"]


class SchemaCollction(enum.Enum):
    """ 当前系统所用Schema文件集合 """
    SCHEMA_ACCOUNTSERVICE = "AccountService.v1_11_0"
    SCHEMA_COMPUTERSYSTEM = "ComputerSystem.v1_18_0"
    SCHEMA_ETHERNETINTERFACE = "EthernetInterface.v1_8_0"
    SCHEMA_ETHERNETINTERFACECOLLECTION = "EthernetInterfaceCollection"
    SCHEMA_JSONSCHEMAFILE = "JsonSchemaFile.v1_1_4"
    SCHEMA_JSONSCHEMAFILECOLLECTION = "JsonSchemaFileCollection"
    SCHEMA_LOGSERVICECOLLECTION = "LogServiceCollection"
    SCHEMA_MANAGERACCOUNT = "ManagerAccount.v1_3_4"
    SCHEMA_MANAGERACCOUNTCOLLECTION = "ManagerAccountCollection"
    SCHEMA_MEMORY = "Memory.v1_15_0"
    SCHEMA_MESSAGEREGISTRY = "MessageRegistry.v1_0_0"
    SCHEMA_MINDXEDGEALARM = "MindXEdgeAlarm.v1_0_0"
    SCHEMA_MINDXEDGEDEVICEINFO = "MindXEdgeDeviceInfo"
    SCHEMA_MINDXEDGEDIGITALWARRANTY = "MindXEdgeDigitalWarranty.v1_0_0"
    SCHEMA_MINDXEDGEETHIPLIST = "MindXEdgeEthIpList.v1_0_0"
    SCHEMA_MINDXEDGEEXTENDEDDEVICE = "MindXEdgeExtendedDevice.v1_0_0"
    SCHEMA_MINDXEDGEHTTPSCERT = "MindXEdgeHttpsCert.v1_0_0"
    SCHEMA_MINDXEDGELTE = "MindXEdgeLTE.v1_0_0"
    SCHEMA_MINDXEDGEMODULECOLLECTION = "MindXEdgeModuleCollection"
    SCHEMA_MINDXEDGENFSMANAGE = "MindXEdgeNfsManage.v1_0_0"
    SCHEMA_MINDXEDGENTPSERVICE = "MindXEdgeNTPService.v1_0_0"
    SCHEMA_MINDXEDGENETMANAGER = "MindXEdgeNetManager.v1_0_0"
    SCHEMA_MINDXEDGEPARTITION = "MindXEdgePartition.v1_0_0"
    SCHEMA_MINDXEDGEPARTITIONCOLLECTION = "MindXEdgePartitionCollection"
    SCHEMA_MINDXEDGESECURITYSERVICE = "MindXEdgeSecurityService.v1_0_0"
    SCHEMA_MINDXEDGESYSTEMTIME = "MindXEdgeSystemTime"
    SCHEMA_NETWORKADAPTERCOLLECTION = "NetworkAdapterCollection"
    SCHEMA_PROCESSOR = "Processor.v1_15_0"
    SCHEMA_PROCESSORCOLLECTION = "ProcessorCollection"
    SCHEMA_SERVICEROOT = "ServiceRoot.v1_9_0"
    SCHEMA_SESSION = "Session.v1_4_0"
    SCHEMA_SESSIONSERVICE = "SessionService.v1_1_8"
    SCHEMA_SIMPLESTORAGE = "SimpleStorage.v1_3_1"
    SCHEMA_SIMPLESTORAGECOLLECTION = "SimpleStorageCollection"
