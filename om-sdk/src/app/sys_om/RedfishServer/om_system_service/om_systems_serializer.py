# Copyright (c) Huawei Technologies Co., Ltd. 2024-2024. All rights reserved.
from ibma_redfish_serializer import Serializer


class DigitalWarrantyResourceSerializer(Serializer):
    """
    功能描述：系统生命周期资源
    """
    service = Serializer.root.systems.digitalwarranty_resource
