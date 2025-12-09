# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
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
