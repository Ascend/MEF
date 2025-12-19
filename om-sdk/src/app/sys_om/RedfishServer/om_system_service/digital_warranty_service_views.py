# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
import json

from flask import request, Blueprint

from common.constants.base_constants import CommonConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from lib_restful_adapter import LibRESTfulAdapter
from om_system_service.om_systems_serializer import DigitalWarrantyResourceSerializer
from system_service.systems_common import make_error_dict
from token_auth import get_privilege_auth

https_digital_warranty_service_bp = Blueprint("DigitalWarrantyService", __name__,
                                              url_prefix="/redfish/v1/Systems/DigitalWarranty")

privilege_auth = get_privilege_auth()


@https_digital_warranty_service_bp.route("", methods=["GET"])
@privilege_auth.token_required
@RedfishGlobals.redfish_operate_adapter(request, "Query DigitalWarranty info")
def service_required():
    """
    服务年限/服务起始时间/出产日期查询
    :return: 响应字典 资源模板或错误消息
    """
    input_err_info = "Get DigitalWarranty info failed."
    try:
        # 获取资源模板
        digitalwarranty_resource = json.loads(DigitalWarrantyResourceSerializer().service.get_resource())
        ret_dict = LibRESTfulAdapter.lib_restful_interface("dflc_info", "GET", None, False)
        return ret_dict, digitalwarranty_resource
    except Exception as err:
        run_log.error("%s reason is:%s", input_err_info, err)
        ret_dict = make_error_dict(input_err_info, CommonConstants.ERR_CODE_400)
        return ret_dict, CommonConstants.ERR_GENERAL_INFO
