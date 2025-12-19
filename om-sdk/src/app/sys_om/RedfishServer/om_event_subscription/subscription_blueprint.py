#!/usr/bin/python3
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from flask import Blueprint
from flask import request

from common.log.logger import run_log
from om_event_subscription.subscription_views import (rf_delete_subscriptions, rf_get_subscriptions,
                                                      rf_create_subscriptions, rf_get_subscriptions_collection,
                                                      rf_event_subscriptions, rf_query_subscription_cert,
                                                      rf_import_root_cert, rf_delete_root_cert, rf_import_root_crl)
from token_auth import get_privilege_auth

privilege_auth = get_privilege_auth()

event_subscription_bp = Blueprint("event_subscription", __name__, url_prefix="/redfish/v1/EventService")

event_subscription_bp.add_url_rule("", view_func=rf_event_subscriptions, methods=["GET"])
event_subscription_bp.add_url_rule("/Subscriptions", view_func=rf_get_subscriptions_collection, methods=["GET"])
event_subscription_bp.add_url_rule("/Subscriptions", view_func=rf_create_subscriptions, methods=["POST"])
event_subscription_bp.add_url_rule("/Subscriptions/<int:subscription_id>",
                                   view_func=rf_get_subscriptions, methods=["GET"])
event_subscription_bp.add_url_rule("/Subscriptions/<int:subscription_id>",
                                   view_func=rf_delete_subscriptions, methods=["DELETE"])

event_subscription_bp.add_url_rule("/ServiceCert", view_func=rf_query_subscription_cert, methods=["GET"])
event_subscription_bp.add_url_rule("/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceRootCA",
                                   view_func=rf_import_root_cert, methods=["POST"])
event_subscription_bp.add_url_rule("/ServiceCert/Actions/ServiceCert."
                                   "DeleteRemoteHttpsServiceRootCA/<int:root_cert_id>",
                                   view_func=rf_delete_root_cert, methods=["DELETE"])
event_subscription_bp.add_url_rule("/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceCrl",
                                   view_func=rf_import_root_crl, methods=["POST"])


@event_subscription_bp.before_request
@privilege_auth.token_required
def set_endpoint_executing_log():
    """event_service进入前先记录一条运行日志"""
    if request.method in ("PATCH", "POST", "DELETE",):
        run_log.info("event subscription service access.")
