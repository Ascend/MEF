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

import json

from flask import Response, make_response

from ibma_redfish_serializer import Serializer


class DefaultResourceMixin:

    def make_default_response(self: Serializer, status: int) -> Response:
        resp = json.loads(self.service.get_resource())
        return make_response(resp, status)


class EventSubscriptionSerializer(Serializer):
    service = Serializer.root.subscription_service.event_resource


class DetailSubscriptionSerializer(Serializer):
    service = Serializer.root.subscription_service.detail_event


class GetSubscriptionCollectionsSerializer(Serializer):
    """
    功能描述：获取订阅资源集合
    """
    service = Serializer.root.subscription_service.get_event_resource


class CreateSubscriptionSerializer(Serializer):
    """
    功能描述：创建订阅资源
    """
    service = Serializer.root.subscription_service.create_event

    def make_response(self, data: dict, status: int):
        resp_json = json.loads(self.service.get_resource())
        resp_json["Id"] = data["id"]
        resp_json["Destination"] = data["Destination"]
        resp_json["EventTypes"] = data["EventTypes"]
        resp_json["Protocol"] = data["Protocol"]
        return {"message": resp_json, "status": status}


class QuerySubscriptionsCertSerializer(Serializer):
    """
    功能描述：查询https根证书
    """
    service = Serializer.root.subscription_service.query_cert
