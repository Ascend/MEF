#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

import json
import threading
from typing import Iterable, NoReturn
from urllib import parse

from flask import request

from cert_manager.parse_tools import ParseCrlInfo
from common.constants.base_constants import CommonConstants
from common.constants.error_codes import SecurityServiceErrorCodes
from common.log.logger import run_log
from ibma_redfish_globals import RedfishGlobals
from ibma_redfish_serializer import SuccessMessageResourceSerializer
from net_manager.checkers.contents_checker import CertContentsChecker
from net_manager.exception import DataCheckException
from net_manager.models import CertManager
from om_event_subscription.constants import MAX_SUBSCRIPTION_ID, MIN_SUBSCRIPTION_ID
from om_event_subscription.models import Subscription, SubscriptionCert
from om_event_subscription.param_checker import (GetOrDelSubscriptionChecker, CreateSubscriptionChecker,
                                                 ImportHttpCertChecker, DeleteHttpRootCertChecker,
                                                 ImportHttpCrlChecker)
from om_event_subscription.subscription_mgr import SubscriptionsMgr, SubscriptionsCertMgr
from om_event_subscription.subscription_serializer import (DetailSubscriptionSerializer, CreateSubscriptionSerializer,
                                                           GetSubscriptionCollectionsSerializer,
                                                           EventSubscriptionSerializer,
                                                           QuerySubscriptionsCertSerializer)
from token_auth import get_privilege_auth

subs_mgr = SubscriptionsMgr()
subs_cert_mgr = SubscriptionsCertMgr()
privilege_auth = get_privilege_auth()

CREATE_SUBSCRIPTION_LOCK = threading.Lock()
DELETE_SUBSCRIPTION_LOCK = threading.Lock()
IMPORT_HTTPS_CRT_LOCK = threading.Lock()
DELETE_HTTPS_CRT_LOCK = threading.Lock()
IMPORT_HTTPS_CRL_LOCK = threading.Lock()


@RedfishGlobals.redfish_operate_adapter(request, "Get Event Service")
def rf_event_subscriptions():
    """
    功能描述：查询事件服务资源
    参数：无
    返回值：响应字典 资源模板或错误消息
    异常描述：NA
    """
    message = "Query Event Service failed."
    ret_dict = {
        "status": CommonConstants.ERR_CODE_200
    }
    try:
        # 获取 EventService资源模板
        resp_json = json.loads(EventSubscriptionSerializer().service.get_resource())
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict["status"] = CommonConstants.ERR_CODE_400
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    run_log.info("Query Event Service successfully.")
    return ret_dict, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Get Event Subscription Collection")
def rf_get_subscriptions_collection():
    """
    功能描述：查询事件订阅集合
    """
    msg = "Get subscription collection success."
    id_list = [str(i) for i in subs_mgr.get_first_subscription()]
    ret_dict = {"status": CommonConstants.ERR_CODE_200, "message": id_list}
    run_log.info(msg)
    try:
        resp_json = json.loads(GetSubscriptionCollectionsSerializer().service.get_resource())
    except Exception as err:
        run_log.error("Get subscription collection failed, reason is: %s", err)
        ret_dict["status"] = CommonConstants.ERR_CODE_400
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    return ret_dict, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Create Event Subscription")
def rf_create_subscriptions():
    """
    功能描述：创建事件订阅
    """
    if CREATE_SUBSCRIPTION_LOCK.locked():
        msg = "Create subscription is busy."
        run_log.error(msg)
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": msg}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with CREATE_SUBSCRIPTION_LOCK:
        msg = "Create subscription failed."
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": msg}
        try:
            request_json = json.loads(parse.unquote(request.get_data().decode("utf-8")))
        except Exception:
            input_err_info = "Request data is not json."
            run_log.error("%s", input_err_info)
            ret_dict["message"] = input_err_info
            return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

        check_ret = RedfishGlobals.check_external_parameter(CreateSubscriptionChecker, request_json)
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO

        if subs_mgr.check_destination_existed(request_json['Destination']):
            input_err_info = "The destination already exists."
            run_log.error("%s", input_err_info)
            ret_dict["message"] = input_err_info
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        if subs_mgr.get_subscription_num() >= MAX_SUBSCRIPTION_ID:
            input_err_info = "The num of subscription exceeds limit."
            run_log.error("%s", input_err_info)
            ret_dict["message"] = input_err_info
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        try:
            min_id = subs_mgr.get_available_min_id()
        except Exception as err:
            run_log.error("Get available min id failed. Because of %s", err)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        request_json["id"] = min_id
        subscription_obj = Subscription.from_dict(request_json)

        try:
            subs_mgr.add_subscription(subscription_obj)
        except Exception as err:
            run_log.error("Create subscription failed. Because of %s", err)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        response = CreateSubscriptionSerializer().make_response(request_json, CommonConstants.ERR_CODE_200)

        try:
            resp_json = json.loads(CreateSubscriptionSerializer().service.get_resource())
        except Exception as err:
            run_log.error("Create subscription failed. Because of %s", err)
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        msg = "Create subscription success."
        run_log.info(msg)
        return response, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Get Event Subscription detail")
def rf_get_subscriptions(subscription_id: int):
    """查询订阅事件资源"""
    check_ret = RedfishGlobals.check_external_parameter(GetOrDelSubscriptionChecker, {"id": subscription_id})
    if check_ret is not None:
        return check_ret, CommonConstants.ERR_GENERAL_INFO
    # 查询详情
    subscription = subs_mgr.subscription_detail(_id=subscription_id)
    if not subscription:
        ret_dict = {"status": CommonConstants.ERR_CODE_404, "message": "SubscriptionDoesNotExists"}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    try:
        resp_json = json.loads(DetailSubscriptionSerializer().service.get_resource())
    except Exception:
        run_log.error("Get subscription failed.")
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": "get subscriptions failed"}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    ret_dict = {"message": subscription, "status": CommonConstants.ERR_CODE_200}
    run_log.info("Query subscription successfully.")
    return ret_dict, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Delete Event Subscription")
def rf_delete_subscriptions(subscription_id: int):
    """删除订阅事件资源"""
    if DELETE_SUBSCRIPTION_LOCK.locked():
        message = "Delete subscription is busy."
        run_log.error(message)
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
        return ret_dict, CommonConstants.ERR_GENERAL_INFO

    with DELETE_SUBSCRIPTION_LOCK:
        check_ret = RedfishGlobals.check_external_parameter(GetOrDelSubscriptionChecker, {"id": subscription_id})
        if check_ret is not None:
            return check_ret, CommonConstants.ERR_GENERAL_INFO
        # 删除资源
        res = subs_mgr.delete_subscription(_id=subscription_id)
        if not res:
            ret_dict = {"status": CommonConstants.ERR_CODE_404, "message": "SubscriptionDoesNotExists"}
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        try:
            resp_json = json.loads(SuccessMessageResourceSerializer().service.get_resource())
        except Exception as err:
            run_log.error("Delete subscription failed. Because of %s", err)
            ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": "Delete subscription failed"}
            return ret_dict, CommonConstants.ERR_GENERAL_INFO

        run_log.info("Delete subscription success.")
        ret_dict = {"status": CommonConstants.ERR_CODE_200}
        return ret_dict, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Query https Cert")
def rf_query_subscription_cert():
    """
    功能描述：查询服务器当前支持的安全服务集合的信息
    参数：无
    返回值：http response
    异常描述：NA
    """
    message = "Query security cert failed."
    resp_json = json.loads(QuerySubscriptionsCertSerializer.service.get_resource())
    try:
        for cert_id in range(MIN_SUBSCRIPTION_ID, MAX_SUBSCRIPTION_ID + 1):
            ret = SubscriptionsCertMgr().get_obj_by_id(cert_id)
            if not ret:
                continue

            resp_json.get("RemoteHttpsServerCertChainInfo").append({"CertId": cert_id, "Usage": "EventSubscription"})
    except Exception as err:
        run_log.error("%s reason is: %s", message, err)
        ret_dict = {
            "status": CommonConstants.ERR_CODE_400,
        }
        return ret_dict, resp_json

    message = "Query security cert success."
    run_log.info(message)

    ret_dict = {"status": CommonConstants.ERR_CODE_200}
    return ret_dict, resp_json


@RedfishGlobals.redfish_operate_adapter(request, "Import Https Root Cert")
def rf_import_root_cert():
    """
    功能描述：导入HTTP传输服务器根证书
    参数：无
    返回值：http response
    异常描述：NA
    """
    message = "Import https root cert failed."
    if IMPORT_HTTPS_CRT_LOCK.locked():
        message = "Import https transmission server root cert failed because lock is locked."
        run_log.error(message)
        ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
        return RedfishGlobals.return_error_info_message(ret_dict, CommonConstants.ERR_GENERAL_INFO)

    with IMPORT_HTTPS_CRT_LOCK:
        try:
            recv_conf_dict = request.get_data()
            # 获取资源模板
            resp_json = json.loads(SuccessMessageResourceSerializer().service.get_resource())

            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                run_log.error("Request data is not json.")
                return ret[1], CommonConstants.ERR_PARTICULAR_INFO

            request_data_dict = ret[1]
            check_ret = ImportHttpCertChecker().check(request_data_dict)
            if not check_ret.success:
                error_response = {"status": CommonConstants.ERR_CODE_400, "message": "Parameter is invalid."}
                run_log.error("Check external parameter failed, Because of %s", check_ret.reason)
                return error_response, CommonConstants.ERR_GENERAL_INFO

            try:
                check_multi_cert(request_data_dict.get("Content"))
            except Exception as err:
                error_response = {"status": CommonConstants.ERR_CODE_400, "message": "Parameter is invalid."}
                run_log.error("Check external parameter failed, Because of %s", str(err))
                return error_response, CommonConstants.ERR_GENERAL_INFO

            subs_cert_dict = {
                "root_cert_id": request_data_dict.get("RootCertId"),
                "type": request_data_dict.get("Type"),
                "cert_contents": request_data_dict.get("Content"),
            }

            try:
                subs_cert_mgr.overwrite_subs_cert(SubscriptionCert.from_dict(subs_cert_dict))
            except Exception as err:
                run_log.error("Import https root cert failed. Because of %s", err)
                ret_dict = {"status": CommonConstants.ERR_CODE_400}
                return ret_dict, CommonConstants.ERR_PARTICULAR_INFO

            message = "Import remote root ca success."
            run_log.info(message)
            ret_dict = {"status": CommonConstants.ERR_CODE_200}
            return ret_dict, resp_json
        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Delete Https Root Cert")
def rf_delete_root_cert(root_cert_id: int):
    """
    功能描述：删除远程HTTPS传输服务器根证书
    参数：root_cert_id 待删除的HTTPS传输服务器根证书号
    :return:http response
    """
    message = "Delete https root cert failed."
    if DELETE_HTTPS_CRT_LOCK.locked():
        error_msg = "Delete https root cert is busy."
        run_log.error(error_msg)
        return RedfishGlobals.make_response(error_msg, CommonConstants.ERR_CODE_500)

    with DELETE_HTTPS_CRT_LOCK:
        try:
            resp_json = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            # 校验参数是否合法
            check_ret = RedfishGlobals.check_external_parameter(DeleteHttpRootCertChecker, {"RootCertId": root_cert_id})
            if check_ret:
                return check_ret, CommonConstants.ERR_GENERAL_INFO

            if not subs_cert_mgr.delete_by_cert_id(root_cert_id):
                message = "https root cert not exist."
                ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
                run_log.error(message)
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            message = "Delete https root cert success."
            run_log.info(message)
            ret_dict = {"status": CommonConstants.ERR_CODE_200}
            return ret_dict, resp_json

        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


@RedfishGlobals.redfish_operate_adapter(request, "Import Https Crl")
def rf_import_root_crl():
    """
    功能描述：导入远程HTTPS传输服务器吊销列表
    参数：无
    :return:http response
    """
    message = "Import https crl failed."
    if IMPORT_HTTPS_CRL_LOCK.locked():
        error_msg = "Import https crl is busy."
        run_log.error(error_msg)
        return RedfishGlobals.make_response(error_msg, CommonConstants.ERR_CODE_500)

    with IMPORT_HTTPS_CRL_LOCK:
        try:
            recv_conf_dict = request.get_data()
            ret = RedfishGlobals.check_json_request(recv_conf_dict)
            if ret[0] != 0:
                run_log.error("Request data is not json.")
                return ret[1], CommonConstants.ERR_PARTICULAR_INFO

            resp_json = json.loads(SuccessMessageResourceSerializer().service.get_resource())
            # 校验参数是否合法
            request_data_dict = ret[1]
            check_ret = ImportHttpCrlChecker().check(request_data_dict)
            if not check_ret.success:
                error_response = {"status": CommonConstants.ERR_CODE_400, "message": "Parameter is invalid."}
                run_log.error("Check external parameter failed, Because of %s", check_ret.reason)
                return error_response, CommonConstants.ERR_GENERAL_INFO

            subs_cert = SubscriptionsCertMgr().get_first()
            if not subs_cert:
                message = "https root cert not been imported."
                run_log.error(message)
                ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
                return ret_dict, CommonConstants.ERR_GENERAL_INFO

            tmp_cert_mgr = CertManager().from_dict({"CertContents": subs_cert.cert_contents})
            crl_contents = request_data_dict.get("Content")
            # 检查该吊销列表是否由对应证书签发
            with ParseCrlInfo(crl_contents) as parse_crl:
                verify_result = parse_crl.verify_crl_buffer_by_ca([tmp_cert_mgr])
                if not verify_result:
                    error_message = "Verify crl against CA certificate failed."
                    run_log.error(error_message)
                    ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
                    return ret_dict, CommonConstants.ERR_GENERAL_INFO

                SubscriptionsCertMgr().update_crt_with_crl(subs_cert.cert_contents, {"crl_contents": crl_contents})
                message = "Import https crl success."
                ret_dict = {"status": CommonConstants.ERR_CODE_200}
                run_log.info(message)
                return ret_dict, resp_json

        except Exception as err:
            run_log.error("%s reason is: %s", message, err)
            ret_dict = {"status": CommonConstants.ERR_CODE_400, "message": message}
            return ret_dict, CommonConstants.ERR_GENERAL_INFO


def node_generator(cert_list: str) -> Iterable[str]:
    sep = "-----END CERTIFICATE-----"
    for node in cert_list.split(sep):
        if not node.split():
            continue
        yield "".join((node, sep))


def check_multi_cert(cert_list: str) -> NoReturn:
    root_ca = 0
    for cert in node_generator(cert_list):
        try:
            check_ret = CertContentsChecker("Content").check({"Content": cert})
        except Exception as err:
            raise DataCheckException(f"cert content checked failed, {err}") from err

        if check_ret.success:
            root_ca += 1
        elif check_ret.err_code != SecurityServiceErrorCodes.ERROR_CERTIFICATE_CA_SIGNATURE_INVALID.code:
            raise DataCheckException(f"cert content checked failed, {check_ret.reason}", err_code=check_ret.err_code)

    if root_ca == 0:
        raise DataCheckException("The certificate chain has no root certificates.")
    elif root_ca > 1:
        raise DataCheckException("The certificate chain contains multiple root certificates.")
