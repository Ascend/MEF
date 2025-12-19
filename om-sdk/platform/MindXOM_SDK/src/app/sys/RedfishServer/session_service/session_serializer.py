# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# OMSDK is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
from ibma_redfish_serializer import Serializer


class SessionServiceResourceSerializer(Serializer):
    """ 功能描述：顶层会话服务资源 """
    service = Serializer.root.sessionService.session_service_resource


class SessionsResourceSerializer(Serializer):
    """ 功能描述：会话集合资源 """
    service = Serializer.root.sessionService.sessions_resource


class SessionsMembersResourceSerializer(Serializer):
    """ 功能描述：指定会话资源 """
    service = Serializer.root.sessionService.sessions_members_resource

