#!/usr/bin/python
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

"""
功    能：Redfish Account Service资源定义
"""

import os.path
from common.ResourceDefV1.resource import RfResource


class RfAccountServiceObj(RfResource):
    """
    功能描述：创建AccountService资源对象
    接口：NA
    """
    ACCOUNT_SERVICE_RESOURCE_DIR = os.path.normpath("redfish/v1/AccountService")
    ACCOUNTS_RESOURCE_DIR = os.path.normpath("redfish/v1/AccountService/Accounts")
    ACCOUNTS_MEMBERS_DIR = os.path.normpath("redfish/v1/AccountService/Accounts/Members")

    account_service_resource: RfResource
    accounts_resource: RfResource
    accounts_members_resource: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.account_service_resource = RfResource(base_path, self.ACCOUNT_SERVICE_RESOURCE_DIR)
        self.accounts_resource = RfResource(base_path, self.ACCOUNTS_RESOURCE_DIR)
        self.accounts_members_resource = RfResource(base_path, self.ACCOUNTS_MEMBERS_DIR)
