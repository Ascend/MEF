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

import os.path

from common.ResourceDefV1.resource import RfResource


class RfEventSubscription(RfResource):
    EVENT_RESOURCE_DIR = os.path.normpath("redfish/v1/EventService")
    GET_EVENT_RESOURCE_DIR = os.path.normpath("redfish/v1/EventService/GetSubscriptions")
    CREATE_EVENT_RESOURCE_DIR = os.path.normpath("redfish/v1/EventService/PostSubscriptions")
    DETAIL_EVENT_RESOURCE_DIR = os.path.normpath("redfish/v1/EventService/GetSubscriptions/Members")
    QUERY_SUBSCRIPTION_CERT = os.path.normpath("redfish/v1/EventService/ServiceCert")

    event_resource: RfResource
    get_event_resource: RfResource
    create_event: RfResource
    detail_event: RfResource
    query_cert: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.event_resource = RfResource(base_path, self.EVENT_RESOURCE_DIR)
        self.get_event_resource = RfResource(base_path, self.GET_EVENT_RESOURCE_DIR)
        self.create_event = RfResource(base_path, self.CREATE_EVENT_RESOURCE_DIR)
        self.detail_event = RfResource(base_path, self.DETAIL_EVENT_RESOURCE_DIR)
        self.query_cert = RfResource(base_path, self.QUERY_SUBSCRIPTION_CERT)
