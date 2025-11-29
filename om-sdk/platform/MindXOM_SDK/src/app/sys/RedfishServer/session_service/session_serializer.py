# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
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

