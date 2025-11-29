# Copyright (c) Huawei Technologies Co., Ltd. 2023-2023. All rights reserved.

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


