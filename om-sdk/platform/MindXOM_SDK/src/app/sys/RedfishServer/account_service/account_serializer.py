# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.
from ibma_redfish_serializer import Serializer


class AccountServiceResourceSerializer(Serializer):
    """
    功能描述：顶层用户服务资源
    """
    service = Serializer.root.AccountService.account_service_resource


class AccountsResourceSerializer(Serializer):
    """
    功能描述：用户集合资源
    """
    service = Serializer.root.AccountService.accounts_resource


class AccountsMembersResourceSerializer(Serializer):
    """
    功能描述：指定用户资源
    """
    service = Serializer.root.AccountService.accounts_members_resource
