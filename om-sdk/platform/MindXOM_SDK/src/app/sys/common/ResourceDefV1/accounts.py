#!/usr/bin/python
# -*- coding: UTF-8 -*-
# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Account Service资源定义
修改记录：2019-1-18 创建
"""

import os.path
from common.ResourceDefV1.resource import RfResource


class RfAccountServiceObj(RfResource):
    """
    功能描述：创建AccountService资源对象
    接口：NA
    修改记录：2022-11-26 创建
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
