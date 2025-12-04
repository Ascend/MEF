# Copyright (c) Huawei Technologies Co., Ltd. 2025-2025. All rights reserved.
# MindEdge is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.

from common.ResourceDefV1.resource import RfResource
from common.constants.product_constants import SERVICE_ROOT


class Serializer:

    root = SERVICE_ROOT

    # 具体序列化类用到的模板服务
    service: RfResource


class SuccessMessageResourceSerializer(Serializer):
    """
    功能描述：成功消息资源
    """
    service = Serializer.root.successmessage.success_message_resource


