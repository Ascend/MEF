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
import os

from flask import request, Blueprint, g

from common.checkers.param_checker import RestoreDefaultConfigChecker
from common.constants.error_codes import CommonErrorCodes
from common.constants.base_constants import CommonConstants, UserManagerConstants, RecoverMiniOSConstants
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from lib_restful_adapter import LibRESTfulAdapter
from om_system_service.default_config import DefaultConfig
from system_service.systems_common import make_error_dict
from token_auth import get_privilege_auth
from user_manager.user_manager import UserManager

om_actions_service_bp = Blueprint("OMActionsService", __name__, url_prefix="/redfish/v1/Systems")

privilege_auth = get_privilege_auth()


@om_actions_service_bp.route("/Actions/RestoreDefaults.Config", methods=["POST"])
@privilege_auth.token_required
@RedfishGlobals.redfish_operate_adapter(request, "Restore default configuration")
def rf_restore_default_configuration():
    """恢复默认配置"""
    if RedfishGlobals.high_risk_exclusive_lock.locked():
        run_log.error("The operation is busy.")
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
            "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, CommonErrorCodes.OPERATE_IS_BUSY.messageKey],
        }
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with RedfishGlobals.high_risk_exclusive_lock:
        ret_dict = LibRESTfulAdapter.lib_restful_interface("ExclusiveStatus", "GET", None, False)
        if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        if not isinstance(ret_dict.get("message"), dict) or not \
                isinstance(ret_dict.get("message").get("system_busy"), bool) or \
                ret_dict.get("message").get("system_busy"):
            run_log.error("The operation is busy.")
            ret_dict = {
                "status": CommonConstants.ERR_CODE_400,
                "message": [CommonErrorCodes.OPERATE_IS_BUSY.code, CommonErrorCodes.OPERATE_IS_BUSY.messageKey],
            }
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        try:
            recv_conf_dict = request.get_data()
            # 校验数据是否为json
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                return ret[1], CommonConstants.ERR_PARTICULAR_INFO

            req_data_dict = ret[1]
            check_ret = RedfishGlobals.check_external_parameter(RestoreDefaultConfigChecker, req_data_dict)
            if check_ret is not None:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            # 二次认证
            req_dict = {
                "oper_type": UserManagerConstants.OPER_TYPE_CHECK_PASSWORD,
                "user_id": g.user_id,
                "password": req_data_dict.get("Password")
            }
            ret_dict = UserManager().get_all_info(req_dict)
            if not LibRESTfulAdapter.check_status_is_ok(ret_dict):
                if RedfishGlobals.get_user_locked_state(ret_dict):
                    message = f"Restore default configuration failed, {CommonConstants.USER_LOCK_STATE}"
                    RedfishGlobals.set_operational_log(request, RedfishGlobals.USER_NAME, message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            ret_dict = DefaultConfig().deal_request(req_data_dict)
            success_message_resource = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            if LibRESTfulAdapter.check_status_is_ok(ret_dict):
                ret_dict = {
                    "status": RedfishGlobals.SuccessfullyCode,
                    "message": {"Message": "Restore defaults configuration successfully."}
                }
                return ret_dict, success_message_resource

            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        except Exception as err:
            message = "Restore defaults configuration failed."
            run_log.error("%s reason is %s", message, err)
            ret_dict = make_error_dict(message, CommonConstants.ERR_CODE_400)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO
