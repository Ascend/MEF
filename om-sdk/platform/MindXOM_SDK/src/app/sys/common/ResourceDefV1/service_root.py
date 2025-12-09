#!/usr/bin/python
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

"""
功    能：Redfish Server 根服务资源定义
修改记录：2016-10-18 创建
"""
import os

from common.ResourceDefV1.public_service import RfPublicServiceObj
from common.ResourceDefV1.accounts import RfAccountServiceObj
from common.ResourceDefV1.errorcolection import RfErrorObj
from common.ResourceDefV1.resource import RfResource
from common.ResourceDefV1.session import RfSessionServiceObj
from common.ResourceDefV1.success_message import RfSuccessMessage
from common.ResourceDefV1.systems import RfSystemsCollection
from common.ResourceDefV1.task_service import RfTaskServiceObj
from common.ResourceDefV1.upgrade_service_common import RfUpgradeService
from common.ResourceDefV1.net_manage_service import RfNetManage


class RfServiceRoot(RfResource):
    """
    功能描述：创建该根服务下的所有资源
    接口：NA
    修改记录：2016-10-18 创建
    """

    class RfOdataServiceDoc(RfResource):
        pass

    def create_sub_objects(self, base_path, rel_path):
        self.create_basic_sub_objects(base_path, rel_path)
        self.create_extend_sub_objects(base_path, rel_path)

    def create_basic_sub_objects(self, base_path, rel_path):
        self.TaskService = RfTaskServiceObj(base_path, os.path.normpath("redfish/v1/TaskService"))
        self.AccountService = RfAccountServiceObj(base_path, os.path.normpath("redfish/v1/AccountService"))
        self.errorcollection = RfErrorObj(base_path, os.path.normpath("redfish/v1/ErrorCollection"))
        self.successmessage = RfSuccessMessage(base_path, os.path.normpath("redfish/v1/SuccessMessage"))
        self.sessionService = RfSessionServiceObj(base_path, os.path.normpath("redfish/v1/SessionService"))
        self.odataServiceDoc = self.RfOdataServiceDoc(base_path, os.path.normpath("redfish/v1/odata"))
        self.upgrade_service_common = RfUpgradeService(base_path, os.path.normpath("redfish/v1/UpgradeService"))
        self.net_manage_service = RfNetManage(base_path, os.path.normpath("redfish/v1/NetManager"))
        self.publicService = RfPublicServiceObj(base_path, os.path.normpath("redfish"))

    def create_extend_sub_objects(self, base_path, rel_path):
        self.systems = RfSystemsCollection(base_path, os.path.normpath("redfish/v1/Systems"))

    def final_init_processing(self, base_path, rel_path):
        pass
