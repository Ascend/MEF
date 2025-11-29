# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

"""
功    能：Redfish Server Systems资源定义
修改记录：2019-1-18 创建
"""

import os
from common.ResourceDefV1.resource import RfResource


class RfSessionServiceObj(RfResource):
    """
    功能描述：创建SessionService资源对象
    接口：NA
    修改记录：2022-12-6 创建
    """

    SESSION_SERVICE_RESOURCE_DIR = os.path.normpath("redfish/v1/SessionService")
    SESSIONS_RESOURCE_DIR = os.path.normpath("redfish/v1/SessionService/Sessions")
    SESSIONS_MEMBERS_DIR = os.path.normpath("redfish/v1/SessionService/Sessions/1")

    session_service_resource: RfResource
    sessions_resource: RfResource
    sessions_members_resource: RfResource

    def create_sub_objects(self, base_path, rel_path):
        self.session_service_resource = RfResource(base_path, self.SESSION_SERVICE_RESOURCE_DIR)
        self.sessions_resource = RfResource(base_path, self.SESSIONS_RESOURCE_DIR)
        self.sessions_members_resource = RfResource(base_path, self.SESSIONS_MEMBERS_DIR)

