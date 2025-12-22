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

import sys
from urllib.parse import urlparse

from common.checkers import ModelChecker, StringChoicesChecker, ListChecker, IntegerChecker
from common.checkers import CheckResult, RegexStringChecker
from common.checkers import AttrCheckerInterface
from common.checkers import PasswordComplexityChecker
from net_manager.checkers.contents_checker import CertContentsChecker, CrlContentsChecker
from om_event_subscription.constants import MIN_SUBSCRIPTION_ID, MAX_SUBSCRIPTION_ID, EventTypes, Protocol, \
    CERT_CRL_FILE_MAX_SIZE_BYTES


class EqualListChecker(ListChecker):
    """
    功能描述：自定义checker。判断列表是否和 contains_list 相等。
    """

    def __init__(
            self,
            attr_name: str = None,
            elem_checker: AttrCheckerInterface = None,
            contains_list: list = None,
            min_len=0,
            max_len=sys.maxsize,
            required: bool = True,
    ):
        super().__init__(attr_name, elem_checker, min_len, max_len, required)
        self.contains_list = contains_list

    def check_dict(self, data: dict) -> CheckResult:
        ret = super().check_dict(data)
        if not ret.success:
            return ret
        value_list = self.raw_value(data)
        if not value_list:
            return CheckResult.make_failed(f"EqualListChecker: {self.name()} is empty.")
        if not isinstance(value_list, (tuple, list)):
            return CheckResult.make_failed(f"EqualListChecker: invalid list type of {self.name()}")
        if self.contains_list == value_list:
            return CheckResult.make_success()
        else:
            return CheckResult.make_failed(f"EqualListChecker: not contains {self.contains_list}")


class GetOrDelSubscriptionChecker(ModelChecker):
    """订阅事件-/redfish/v1/EventService/Subscriptions-GET、DELETE"""

    class Meta:
        fields = (
            IntegerChecker("id", min_value=MIN_SUBSCRIPTION_ID, max_value=MAX_SUBSCRIPTION_ID),
        )


class HttpHeadersChecker(ModelChecker):
    """
    功能描述：校验 Http Headers
    """

    class Meta:
        fields = (
            PasswordComplexityChecker("X-Auth-Token", min_len=24, max_len=128),
        )


class DestinationChecker(RegexStringChecker):
    def __init__(self, attr_name: str = None, min_len: int = 0, max_len: int = 2048, required: bool = True):
        super().__init__(attr_name, "(https|HTTPS)://[-A-Za-z0-9+&#/%?=~_|!:,.;]*[-A-Za-z0-9+&#/%=~_|]", min_len,
                         max_len, required)

    def check_dict(self, data: dict) -> CheckResult:
        result = super().check_dict(data)
        if not result.success:
            return result
        value = self.raw_value(data)
        try:
            parse_ret = urlparse(value)
        except Exception:
            msg_format = "DestinationChecker: parse destination of {} failed."
            return CheckResult.make_failed(msg_format.format(self.name()))

        if parse_ret.username or parse_ret.password:
            return CheckResult.make_failed("Destination is invalid")

        return CheckResult.make_success()


class CreateSubscriptionChecker(ModelChecker):
    """订阅事件-/redfish/v1/EventService/Subscriptions-POST"""

    class Meta:
        fields = (
            DestinationChecker("Destination"),
            EqualListChecker("EventTypes", contains_list=EventTypes.get_supported_types()),
            HttpHeadersChecker("HttpHeaders"),
            StringChoicesChecker("Protocol", choices=(Protocol.REDFISH.value,)),
        )


class ImportHttpCertChecker(ModelChecker):
    """接口支持导入HTTPS传输服务器根证书-/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceRootCA-POST
    """

    class Meta:
        fields = (
            StringChoicesChecker("Type", choices=("text",)),
            CertContentsChecker("Content", min_len=1, max_len=CERT_CRL_FILE_MAX_SIZE_BYTES),
            IntegerChecker("RootCertId", min_value=MIN_SUBSCRIPTION_ID, max_value=MAX_SUBSCRIPTION_ID, restrict=True),
        )


class DeleteHttpRootCertChecker(ModelChecker):
    """删除远程HTTPS传输服务器根证书-/ServiceCert/Actions/ServiceCert/DeleteRemoteHttpsServiceRootCA/<int:RootCertId>-DELETE
    """

    class Meta:
        fields = (
            IntegerChecker("RootCertId", min_value=MIN_SUBSCRIPTION_ID, max_value=MAX_SUBSCRIPTION_ID, restrict=True),
        )


class ImportHttpCrlChecker(ModelChecker):
    """导入远程HTTPS传输服务器根证书的吊销列表-
    /redfish/v1/EdgeSystem/SecurityService/ServiceCert/Actions/ServiceCert.ImportRemoteHttpsServiceCrl-POST
    """

    class Meta:
        fields = (
            StringChoicesChecker("Type", choices=("text",)),
            CrlContentsChecker("Content", min_len=1, max_len=CERT_CRL_FILE_MAX_SIZE_BYTES),
            IntegerChecker("RootCertId", min_value=MIN_SUBSCRIPTION_ID, max_value=MAX_SUBSCRIPTION_ID, restrict=True),
        )
